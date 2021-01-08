// Copyright 2014,2015,2016,2017,2018,2019,2020,2021 SeukWon Kang (kasworld@gmail.com)
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gorillaloop

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kasworld/mininet/lib/packet"
)

func SendControl(
	wsConn *websocket.Conn, mt int, PacketWriteTimeOut time.Duration) error {

	return wsConn.WriteControl(mt, []byte{}, time.Now().Add(PacketWriteTimeOut))
}

func WritePacket(wsConn *websocket.Conn, pk *packet.Packet) error {
	w, err := wsConn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	if err := packet.WritePacket(w, pk); err != nil {
		return err
	}
	return w.Close()
}

func SendLoop(
	sendRecvCtx context.Context,
	SendRecvStop func(),
	wsConn *websocket.Conn,
	timeout time.Duration,
	SendCh chan *packet.Packet,
	handleSentPacketFn func(pk *packet.Packet) error,
) error {

	defer SendRecvStop()
	var err error
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			err = SendControl(wsConn, websocket.CloseMessage, timeout)
			break loop
		case pk := <-SendCh:
			if err = wsConn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
				break loop
			}
			if err = WritePacket(wsConn, pk); err != nil {
				break loop
			}
			if err = handleSentPacketFn(pk); err != nil {
				break loop
			}
		}
	}
	return err
}

func RecvLoop(
	sendRecvCtx context.Context,
	SendRecvStop func(),
	wsConn *websocket.Conn,
	timeout time.Duration,
	HandleRecvPacketFn func(pk *packet.Packet) error,
) error {
	defer SendRecvStop()
	var err error
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			break loop
		default:
			if err = wsConn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				break loop
			}
			if pk, lerr := RecvPacket(wsConn); lerr != nil {
				if operr, ok := lerr.(*net.OpError); ok && operr.Timeout() {
					continue
				}
				err = lerr
				break loop
			} else {
				if err = HandleRecvPacketFn(pk); err != nil {
					break loop
				}
			}
		}
	}
	return err
}

func RecvPacket(wsConn *websocket.Conn) (*packet.Packet, error) {
	mt, rdata, err := wsConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if mt != websocket.BinaryMessage {
		return nil, fmt.Errorf("message not binary %v", mt)
	}
	return packet.Bytes2Packet(rdata)
}
