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

package minitcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kasworld/mininet/packet"
)

type Connection struct {
	conn         *net.TCPConn
	sendCh       chan *packet.Packet
	sendRecvStop func()

	readTimeoutSec     time.Duration
	writeTimeoutSec    time.Duration
	handleRecvPacketFn func(pk *packet.Packet) error
	handleSentPacketFn func(pk *packet.Packet) error
}

func New(
	readTimeoutSec, writeTimeoutSec time.Duration,
	handleRecvPacketFn func(pk *packet.Packet) error,
	handleSentPacketFn func(pk *packet.Packet) error,
) *Connection {
	tc := &Connection{
		sendCh:             make(chan *packet.Packet, 10),
		readTimeoutSec:     readTimeoutSec,
		writeTimeoutSec:    writeTimeoutSec,
		handleRecvPacketFn: handleRecvPacketFn,
		handleSentPacketFn: handleSentPacketFn,
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

func (tc *Connection) Run(mainctx context.Context) error {
	sendRecvCtx, sendRecvCancel := context.WithCancel(mainctx)
	tc.sendRecvStop = sendRecvCancel
	var rtnerr error
	var sendRecvWaitGroup sync.WaitGroup
	sendRecvWaitGroup.Add(2)
	go func() {
		defer sendRecvWaitGroup.Done()
		err := RecvLoop(
			sendRecvCtx,
			tc.sendRecvStop,
			tc.conn,
			tc.readTimeoutSec,
			tc.handleRecvPacketFn)
		if err != nil {
			rtnerr = err
		}
	}()
	go func() {
		defer sendRecvWaitGroup.Done()
		err := SendLoop(
			sendRecvCtx,
			tc.sendRecvStop,
			tc.conn,
			tc.writeTimeoutSec,
			tc.sendCh,
			tc.handleSentPacketFn)
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

func SendLoop(sendRecvCtx context.Context, SendRecvStop func(), tcpConn *net.TCPConn,
	timeOut time.Duration,
	SendCh chan *packet.Packet,
	handleSentPacketFn func(pk *packet.Packet) error,
) error {

	defer SendRecvStop()
	var err error
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			break loop
		case pk := <-SendCh:
			if err = tcpConn.SetWriteDeadline(time.Now().Add(timeOut)); err != nil {
				break loop
			}
			if err = packet.WritePacket(tcpConn, pk); err != nil {
				break loop
			}
			if err = handleSentPacketFn(pk); err != nil {
				break loop
			}
		}
	}
	return err
}

func RecvLoop(sendRecvCtx context.Context, SendRecvStop func(), tcpConn *net.TCPConn,
	timeOut time.Duration,
	HandleRecvPacketFn func(pk *packet.Packet) error,
) error {

	defer SendRecvStop()

	var err error
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			return nil

		default:
			if err = tcpConn.SetReadDeadline(time.Now().Add(timeOut)); err != nil {
				break loop
			}
			pk, err := packet.ReadPacket(tcpConn)
			if err != nil {
				return err
			}
			if err = HandleRecvPacketFn(pk); err != nil {
				break loop
			}
		}
	}
	return err
}
