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

package tcploop

import (
	"context"
	"net"
	"time"

	"github.com/kasworld/mininet/lib/packet"
)

func SendLoop(
	sendRecvCtx context.Context,
	SendRecvStop func(),
	tcpConn *net.TCPConn,
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

func RecvLoop(
	sendRecvCtx context.Context,
	SendRecvStop func(),
	tcpConn *net.TCPConn,
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
