// Copyright 2014,2015,2016,2017,2018,2019,2020 SeukWon Kang (kasworld@gmail.com)
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serveconn

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kasworld/mininet/gorillaloop"
	"github.com/kasworld/mininet/packet"
	"github.com/kasworld/mininet/tcploop"
)

func (scb *ServeConn) String() string {
	return fmt.Sprintf("ServeConn[SendCh:%v/%v]",
		len(scb.sendCh), cap(scb.sendCh))
}

type ServeConn struct {
	connData     interface{} // custom data for this conn
	sendCh       chan *packet.Packet
	sendRecvStop func()
}

// New with stats local
func New(connData interface{}, sendBufferSize int) *ServeConn {
	scb := &ServeConn{
		connData: connData,
		sendCh:   make(chan *packet.Packet, sendBufferSize),
	}
	scb.sendRecvStop = func() {
		fmt.Printf("Too early sendRecvStop call %v\n", scb)
	}
	return scb
}

func (scb *ServeConn) Disconnect() {
	scb.sendRecvStop()
}
func (scb *ServeConn) GetConnData() interface{} {
	return scb.connData
}
func (scb *ServeConn) EnqueueSendPacket(pk *packet.Packet) error {
	select {
	case scb.sendCh <- pk:
		return nil
	default:
		return fmt.Errorf("Send channel full %v", scb)
	}
}

func (scb *ServeConn) ServeTCP(
	mainctx context.Context, conn *net.TCPConn,
	readTimeoutSec, writeTimeoutSec time.Duration,
	handleRecvPacketFn func(me interface{}, pk *packet.Packet) error,
	handleSentPacketFn func(me interface{}, pk *packet.Packet) error,
) error {
	var returnerr error
	sendRecvCtx, sendRecvCancel := context.WithCancel(mainctx)
	scb.sendRecvStop = sendRecvCancel
	go func() {
		err := tcploop.RecvLoop(sendRecvCtx, scb.sendRecvStop, conn,
			readTimeoutSec,
			func(pk *packet.Packet) error { // add connection info
				return handleRecvPacketFn(scb, pk)
			},
		)
		if err != nil {
			returnerr = fmt.Errorf("end RecvLoop %v", err)
		}
	}()
	go func() {
		err := tcploop.SendLoop(sendRecvCtx, scb.sendRecvStop, conn,
			writeTimeoutSec, scb.sendCh,
			func(pk *packet.Packet) error { // add connection info
				return handleSentPacketFn(scb, pk)
			},
		)
		if err != nil {
			returnerr = fmt.Errorf("end SendLoop %v", err)
		}
	}()
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			break loop
		}
	}
	return returnerr
}

func (scb *ServeConn) ServeWS(
	mainctx context.Context, conn *websocket.Conn,
	readTimeoutSec, writeTimeoutSec time.Duration,
	handleRecvPacketFn func(me interface{}, pk *packet.Packet) error,
	handleSentPacketFn func(me interface{}, pk *packet.Packet) error,
) error {
	var returnerr error
	sendRecvCtx, sendRecvCancel := context.WithCancel(mainctx)
	scb.sendRecvStop = sendRecvCancel
	go func() {
		err := gorillaloop.RecvLoop(sendRecvCtx, scb.sendRecvStop, conn,
			readTimeoutSec,
			func(pk *packet.Packet) error { // add connection info
				return handleRecvPacketFn(scb, pk)
			},
		)
		if err != nil {
			returnerr = fmt.Errorf("end RecvLoop %v", err)
		}
	}()
	go func() {
		err := gorillaloop.SendLoop(sendRecvCtx, scb.sendRecvStop, conn,
			writeTimeoutSec, scb.sendCh,
			func(pk *packet.Packet) error { // add connection info
				return handleSentPacketFn(scb, pk)
			},
		)
		if err != nil {
			returnerr = fmt.Errorf("end SendLoop %v", err)
		}
	}()
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			break loop
		}
	}
	return returnerr
}
