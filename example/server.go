// Copyright 2015,2016,2017,2018,2019,2020 SeukWon Kang (kasworld@gmail.com)
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kasworld/actpersec"
	"github.com/kasworld/argdefault"
	"github.com/kasworld/mininet/example/enum/commandid"
	"github.com/kasworld/mininet/example/enum/flowtype"
	"github.com/kasworld/mininet/example/enum/resultcode"
	"github.com/kasworld/mininet/example/idu64str"
	"github.com/kasworld/mininet/example/netobj"
	"github.com/kasworld/mininet/lib/connmanager"
	"github.com/kasworld/mininet/lib/header"
	"github.com/kasworld/mininet/lib/packet"
	"github.com/kasworld/mininet/lib/serveconn"
	"github.com/kasworld/prettystring"
)

// service const
const (
	sendBufferSize  = 10
	readTimeoutSec  = 6 * time.Second
	writeTimeoutSec = 3 * time.Second
)

type ServerConfig struct {
	TcpPort    string `default:":8081" argname:""`
	HttpPort   string `default:":8080" argname:""`
	HttpFolder string `default:"www" argname:""`
}

func main() {
	ads := argdefault.New(&ServerConfig{})
	ads.RegisterFlag()
	flag.Parse()
	config := &ServerConfig{}
	ads.SetDefaultToNonZeroField(config)
	ads.ApplyFlagTo(config)
	fmt.Println(prettystring.PrettyString(config, 4))

	svr := NewServer(config)
	svr.Run()
}

type Server struct {
	config       *ServerConfig
	sendRecvStop func()

	connManager *connmanager.Manager

	SendStat *actpersec.ActPerSec `prettystring:"simple"`
	RecvStat *actpersec.ActPerSec `prettystring:"simple"`

	DemuxReq2BytesAPIFnMap [commandid.CommandID_Count]func(
		me interface{}, pk *packet.Packet) (
		header.Header, interface{}, error)
}

func NewServer(config *ServerConfig) *Server {
	svr := &Server{
		config:      config,
		connManager: connmanager.New(),
		SendStat:    actpersec.New(),
		RecvStat:    actpersec.New(),
	}
	svr.sendRecvStop = func() {
		fmt.Printf("Too early sendRecvStop call\n")
	}

	svr.DemuxReq2BytesAPIFnMap = [...]func(
		me interface{}, pk *packet.Packet) (
		header.Header, interface{}, error){
		commandid.Invalid: svr.bytesAPIFn_ReqInvalid,
		commandid.Echo:    svr.bytesAPIFn_ReqEcho,
	}
	return svr
}

func (svr *Server) Run() {
	ctx, stopFn := context.WithCancel(context.Background())
	svr.sendRecvStop = stopFn
	defer svr.sendRecvStop()

	go svr.serveTCP(ctx, svr.config.TcpPort)
	go svr.serveHTTP(ctx, svr.config.HttpPort, svr.config.HttpFolder)

	timerInfoTk := time.NewTicker(1 * time.Second)
	defer timerInfoTk.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timerInfoTk.C:
			svr.SendStat.UpdateLap()
			svr.RecvStat.UpdateLap()
			fmt.Printf("Connection:%v Send:%v Recv:%v\n",
				svr.connManager.Len(),
				svr.SendStat, svr.RecvStat)
		}
	}
}

func (svr *Server) serveHTTP(ctx context.Context, port string, folder string) {
	fmt.Printf("http server dir=%v port=%v , http://localhost%v/\n",
		folder, port, port)
	webMux := http.NewServeMux()
	webMux.Handle("/",
		http.FileServer(http.Dir(folder)),
	)
	webMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		svr.serveWebSocketClient(ctx, w, r)
	})
	if err := http.ListenAndServe(port, webMux); err != nil {
		fmt.Println(err.Error())
	}
}

func CheckOrigin(r *http.Request) bool {
	return true
}

func (svr *Server) serveWebSocketClient(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: CheckOrigin,
	}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("upgrade %v\n", err)
		return
	}

	connID := idu64str.G_Maker.New()
	c2sc := serveconn.New(connID, sendBufferSize)

	// add to conn manager
	svr.connManager.Add(connID, c2sc)

	// start client service
	c2sc.ServeWS(ctx, wsConn,
		readTimeoutSec, writeTimeoutSec,
		svr.handleRecvPacketFn, svr.handleSentPacketFn,
	)

	// connection cleanup here
	wsConn.Close()

	// del from conn manager
	svr.connManager.Del(connID)
}

func (svr *Server) serveTCP(ctx context.Context, port string) {
	fmt.Printf("tcp server port=%v\n", port)
	tcpaddr, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}
	listener, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}
	defer listener.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			listener.SetDeadline(time.Now().Add(time.Duration(1 * time.Second)))
			conn, err := listener.AcceptTCP()
			if err != nil {
				operr, ok := err.(*net.OpError)
				if ok && operr.Timeout() {
					continue
				}
				fmt.Printf("error %#v\n", err)
			} else {
				go svr.serveTCPClient(ctx, conn)
			}
		}
	}
}

func (svr *Server) serveTCPClient(ctx context.Context, conn *net.TCPConn) {

	connID := idu64str.G_Maker.New()

	c2sc := serveconn.New(connID, sendBufferSize)

	// add to conn manager
	svr.connManager.Add(connID, c2sc)

	// start client service
	c2sc.ServeTCP(ctx, conn,
		readTimeoutSec, writeTimeoutSec,
		svr.handleRecvPacketFn, svr.handleSentPacketFn,
	)

	// connection cleanup here
	c2sc.Disconnect()
	conn.Close()

	// del from conn manager
	svr.connManager.Del(connID)
}

///////////////////////////////////////////////////////////////

func (svr *Server) handleRecvPacketFn(me interface{}, pk *packet.Packet) error {
	svr.RecvStat.Inc()
	switch flowtype.FlowType(pk.Header.FlowType) {
	default:
		return fmt.Errorf("Invalid packet type %v", pk.Header)
	case flowtype.Request:
		conn := me.(*serveconn.ServeConn)
		fn := svr.DemuxReq2BytesAPIFnMap[pk.Header.CommandID]
		hd, body, err := fn(me, pk)
		if err != nil {
			fmt.Printf("%v\n", err)
			return err
		}
		bodybytes, _, err := netobj.MarshalBody_msgp(body, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
			return err
		}
		shd := hd
		shd.FlowType = uint8(flowtype.Response)
		shd.BodyLen = uint32(len(bodybytes))
		spk := &packet.Packet{
			Header: shd,
			Body:   bodybytes,
		}
		if err := conn.EnqueueSendPacket(spk); err != nil {
			fmt.Printf("End %v %v %v\n", conn, spk, err)
			conn.Disconnect()
			return fmt.Errorf("Send fail %v %v", conn, err)
		}
	}
	return nil
}
func (svr *Server) handleSentPacketFn(me interface{}, pk *packet.Packet) error {
	svr.SendStat.Inc()
	switch flowtype.FlowType(pk.Header.FlowType) {
	default:
		return fmt.Errorf("Invalid packet type %v", pk.Header)
	case flowtype.Notification:
	case flowtype.Response:
	}
	return nil
}

///////////////////////////////////////////////////////////////

func (svr *Server) bytesAPIFn_ReqInvalid(
	me interface{}, pk *packet.Packet) (
	header.Header, interface{}, error) {
	return header.Header{}, nil, fmt.Errorf("invalid packet")
}

func (svr *Server) bytesAPIFn_ReqEcho(
	me interface{}, pk *packet.Packet) (
	header.Header, interface{}, error) {

	robj, err := netobj.Unmarshal_ReqEcho_msgp(pk)
	if err != nil {
		return pk.Header, nil, fmt.Errorf("unmarshal_ReqEcho %v", err)
	}
	recvBody, ok := robj.(*netobj.ReqEcho_data)
	if !ok {
		return pk.Header, nil, fmt.Errorf("Packet type miss match %v", robj)
	}
	_ = recvBody

	shd := pk.Header
	shd.ResultCode = uint16(resultcode.None)
	sendBody := &netobj.RspEcho_data{recvBody.Msg}
	return shd, sendBody, nil
}
