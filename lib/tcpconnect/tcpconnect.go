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

package tcpconnect

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kasworld/mininet/lib/packet"
	"github.com/kasworld/mininet/lib/tcploop"
)

type Connection struct {
	conn         *net.TCPConn
	sendCh       chan *packet.Packet
	sendRecvStop func()
}

func New(sendBufferSize int) *Connection {
	tc := &Connection{
		sendCh: make(chan *packet.Packet, sendBufferSize),
	}

	tc.sendRecvStop = func() {
		fmt.Printf("Too early sendRecvStop call %v\n", tc)
	}
	return tc
}

func (tc *Connection) ConnectTo(remoteAddr string) error {
	tcpaddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return err
	}
	tc.conn, err = net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return err
	}
	return nil
}

func (tc *Connection) Cleanup() {
	tc.sendRecvStop()
	if tc.conn != nil {
		tc.conn.Close()
	}
}

func (tc *Connection) Run(
	mainctx context.Context,
	readTimeoutSec, writeTimeoutSec time.Duration,
	handleRecvPacketFn func(pk *packet.Packet) error,
	handleSentPacketFn func(pk *packet.Packet) error,
) error {
	sendRecvCtx, sendRecvCancel := context.WithCancel(mainctx)
	tc.sendRecvStop = sendRecvCancel
	var rtnerr error
	var sendRecvWaitGroup sync.WaitGroup
	sendRecvWaitGroup.Add(2)
	go func() {
		defer sendRecvWaitGroup.Done()
		err := tcploop.RecvLoop(
			sendRecvCtx,
			tc.sendRecvStop,
			tc.conn,
			readTimeoutSec,
			handleRecvPacketFn)
		if err != nil {
			rtnerr = err
		}
	}()
	go func() {
		defer sendRecvWaitGroup.Done()
		err := tcploop.SendLoop(
			sendRecvCtx,
			tc.sendRecvStop,
			tc.conn,
			writeTimeoutSec,
			tc.sendCh,
			handleSentPacketFn)
		if err != nil {
			rtnerr = err
		}
	}()
	sendRecvWaitGroup.Wait()
	return rtnerr
}

func (tc *Connection) EnqueueSendPacket(pk *packet.Packet) error {
	select {
	case tc.sendCh <- pk:
		return nil
	default:
		return fmt.Errorf("Send channel full %v", tc)
	}
}
