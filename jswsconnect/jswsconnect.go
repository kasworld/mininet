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

package jswsconnect

import (
	"context"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/kasworld/mininet/packet"
)

type Connection struct {
	remoteAddr   string
	conn         js.Value
	SendRecvStop func()
	sendCh       chan *packet.Packet

	handleRecvPacketFn func(pk *packet.Packet) error
	handleSentPacketFn func(pk *packet.Packet) error
}

func (wsc *Connection) String() string {
	return fmt.Sprintf("Connection[%v SendCh:%v]",
		wsc.remoteAddr, len(wsc.sendCh))
}

func New(
	connAddr string,
	handleRecvPacketFn func(pk *packet.Packet) error,
	handleSentPacketFn func(pk *packet.Packet) error,
) *Connection {
	wsc := &Connection{
		remoteAddr:         connAddr,
		sendCh:             make(chan *packet.Packet, 10),
		handleRecvPacketFn: handleRecvPacketFn,
		handleSentPacketFn: handleSentPacketFn,
	}
	wsc.SendRecvStop = func() {
		JsLogErrorf("Too early SendRecvStop call %v", wsc)
	}
	return wsc
}

func (wsc *Connection) Connect(
	ctx context.Context,
	wg *sync.WaitGroup,
) error {
	connCtx, ctxCancel := context.WithCancel(ctx)
	wsc.SendRecvStop = ctxCancel

	wsc.conn = js.Global().Get("WebSocket").New(wsc.remoteAddr)
	if !wsc.conn.Truthy() {
		err := fmt.Errorf("fail to connect %v", wsc.remoteAddr)
		JsLogErrorf("%v", err)
		return err
	}
	wsc.conn.Call("addEventListener", "open", js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			wsc.conn.Call("addEventListener", "message", js.FuncOf(wsc.handleWebsocketMessage))
			go wsc.sendLoop(connCtx)
			wg.Done()
			return nil
		}))
	wsc.conn.Call("addEventListener", "close", js.FuncOf(wsc.wsClosed))
	wsc.conn.Call("addEventListener", "error", js.FuncOf(wsc.wsError))
	return nil
}

func (wsc *Connection) wsClosed(this js.Value, args []js.Value) interface{} {
	wsc.SendRecvStop()
	JsLogError("ws closed")
	return nil
}

func (wsc *Connection) wsError(this js.Value, args []js.Value) interface{} {
	wsc.SendRecvStop()
	JsLogError(this, args)
	return nil
}

func (wsc *Connection) sendLoop(sendRecvCtx context.Context) {
	defer wsc.SendRecvStop()
	var err error
loop:
	for {
		select {
		case <-sendRecvCtx.Done():
			break loop
		case pk := <-wsc.sendCh:
			if err = wsc.sendPacket(pk); err != nil {
				break loop
			}
			if err = wsc.handleSentPacketFn(pk); err != nil {
				break loop
			}
		}
	}
	JsLogErrorf("end SendLoop %v\n", err)
	return
}

func (wsc *Connection) sendPacket(pk *packet.Packet) error {
	sendBuffer := append(pk.Header.ToByteList(), pk.Body...)
	sendData := js.Global().Get("Uint8Array").New(len(sendBuffer))
	js.CopyBytesToJS(sendData, sendBuffer)
	wsc.conn.Call("send", sendData)
	return nil
}

func (wsc *Connection) handleWebsocketMessage(this js.Value, args []js.Value) interface{} {
	data := args[0].Get("data") // blob
	aBuff := data.Call("arrayBuffer")
	aBuff.Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			rdata := ArrayBufferToSlice(args[0])
			pk, lerr := packet.Bytes2Packet(rdata)
			if lerr != nil {
				JsLogError(lerr.Error())
				wsc.SendRecvStop()
				return nil
			} else {
				if err := wsc.handleRecvPacketFn(pk); err != nil {
					JsLogErrorf("%v", err)
					wsc.SendRecvStop()
					return nil
				}
			}
			return nil
		}))

	return nil
}

func Uint8ArrayToSlice(value js.Value) []byte {
	s := make([]byte, value.Get("byteLength").Int())
	js.CopyBytesToGo(s, value)
	return s
}

func ArrayBufferToSlice(value js.Value) []byte {
	return Uint8ArrayToSlice(js.Global().Get("Uint8Array").New(value))
}

func (wsc *Connection) EnqueueSendPacket(pk packet.Packet) error {
	select {
	case wsc.sendCh <- pk:
		return nil
	default:
		return fmt.Errorf("Send channel full %v", wsc)
	}
}

/////////

func JsLogError(v ...interface{}) {
	js.Global().Get("console").Call("error", v...)
}

func JsLogErrorf(format string, v ...interface{}) {
	js.Global().Get("console").Call("error", fmt.Sprintf(format, v...))
}
