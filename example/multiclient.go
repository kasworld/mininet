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
	"time"

	"github.com/kasworld/argdefault"
	"github.com/kasworld/mininet/example/enum/commandid"
	"github.com/kasworld/mininet/example/enum/flowtype"
	"github.com/kasworld/mininet/example/netobj"
	"github.com/kasworld/mininet/lib/gorillaconnect"
	"github.com/kasworld/mininet/lib/header"
	"github.com/kasworld/mininet/lib/packet"
	"github.com/kasworld/mininet/lib/packetid2rspfn"
	"github.com/kasworld/mininet/lib/tcpconnect"
	"github.com/kasworld/multirun"
	"github.com/kasworld/prettystring"
	"github.com/kasworld/rangestat"
)

// service const
const (
	// for client
	readTimeoutSec  = 6 * time.Second
	writeTimeoutSec = 3 * time.Second
	sendBufferSize  = 10
)

type MultiClientConfig struct {
	ConnectToServer  string `default:"localhost:8080" argname:""`
	NetType          string `default:"ws" argname:""`
	PlayerNameBase   string `default:"MC_" argname:""`
	Concurrent       int    `default:"50000" argname:""`
	AccountPool      int    `default:"0" argname:""`
	AccountOverlap   int    `default:"0" argname:""`
	LimitStartCount  int    `default:"0" argname:""`
	LimitEndCount    int    `default:"0" argname:""`
	PacketIntervalMS int    `default:"1000" argname:""` // milisecond
}

func main() {
	ads := argdefault.New(&MultiClientConfig{})
	ads.RegisterFlag()
	flag.Parse()
	config := &MultiClientConfig{}
	ads.SetDefaultToNonZeroField(config)
	ads.ApplyFlagTo(config)
	fmt.Println(prettystring.PrettyString(config, 4))

	chErr := make(chan error)
	go multirun.Run(
		context.Background(),
		config.Concurrent,
		config.AccountPool,
		config.AccountOverlap,
		config.LimitStartCount,
		config.LimitEndCount,
		func(config interface{}) multirun.ClientI {
			return NewApp(config.(AppArg))
		},
		func(i int) interface{} {
			return AppArg{
				ConnectToServer:  config.ConnectToServer,
				NetType:          config.NetType,
				Nickname:         fmt.Sprintf("%v%v", config.PlayerNameBase, i),
				SessionUUID:      "",
				Auth:             "",
				PacketIntervalMS: config.PacketIntervalMS,
			}
		},
		chErr,
		rangestat.New("", 0, config.Concurrent),
	)
	for err := range chErr {
		fmt.Printf("%v\n", err)
	}
}

type AppArg struct {
	ConnectToServer  string
	NetType          string
	Nickname         string
	SessionUUID      string
	Auth             string
	PacketIntervalMS int
}

type App struct {
	config            AppArg
	c2scWS            *gorillaconnect.Connection
	c2scTCP           *tcpconnect.Connection
	EnqueueSendPacket func(pk *packet.Packet) error
	runResult         error

	sendRecvStop func()
	pid2recv     *packetid2rspfn.PID2RspFn
}

func NewApp(config AppArg) *App {
	app := &App{
		config:   config,
		pid2recv: packetid2rspfn.New(),
	}
	return app
}

func (app *App) String() string {
	return fmt.Sprintf("App[%v %v]", app.config.Nickname, app.config.SessionUUID)
}

func (app *App) GetArg() interface{} {
	return app.config
}

func (app *App) GetRunResult() error {
	return app.runResult
}

func (app *App) Run(mainctx context.Context) {
	ctx, stopFn := context.WithCancel(mainctx)
	app.sendRecvStop = stopFn
	defer app.sendRecvStop()

	switch app.config.NetType {
	default:
		fmt.Printf("unsupported nettype %v\n", app.config.NetType)
		return
	case "tcp":
		app.connectTCP(ctx)
	case "ws":
		app.connectWS(ctx)
	}

	if app.config.PacketIntervalMS == 0 {
		go app.reqEcho()
		for {
			select {
			case <-ctx.Done():
				return
			}
		}
	} else {
		timerPingTk := time.NewTicker(time.Millisecond * time.Duration(app.config.PacketIntervalMS))
		defer timerPingTk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timerPingTk.C:
				go app.reqEcho()

			}
		}
	}

}

func (app *App) connectWS(ctx context.Context) {
	app.c2scWS = gorillaconnect.New(sendBufferSize)
	if err := app.c2scWS.ConnectTo(app.config.ConnectToServer); err != nil {
		app.runResult = err
		fmt.Printf("%v\n", err)
		app.sendRecvStop()
		return
	}
	app.EnqueueSendPacket = app.c2scWS.EnqueueSendPacket
	go func(ctx context.Context) {
		app.runResult = app.c2scWS.Run(ctx,
			readTimeoutSec, writeTimeoutSec,
			app.handleRecvPacket,
			app.handleSentPacket,
		)
	}(ctx)
}

func (app *App) connectTCP(ctx context.Context) {
	app.c2scTCP = tcpconnect.New(sendBufferSize)
	if err := app.c2scTCP.ConnectTo(app.config.ConnectToServer); err != nil {
		app.runResult = err
		fmt.Printf("%v\n", err)
		app.sendRecvStop()
		return
	}
	app.EnqueueSendPacket = app.c2scTCP.EnqueueSendPacket
	go func(ctx context.Context) {
		app.runResult = app.c2scTCP.Run(ctx,
			readTimeoutSec, writeTimeoutSec,
			app.handleRecvPacket,
			app.handleSentPacket,
		)
	}(ctx)
}

func (app *App) reqEcho() error {
	msg := ""
	// msg := fmt.Sprintf("hello world from %v", app.config.Nickname)
	return app.ReqWithRspFn(
		commandid.Echo,
		&netobj.ReqEcho_data{Msg: msg},
		func(pk *packet.Packet) error {
			go app.reqEcho()
			return nil
		},
	)
}

func (app *App) handleSentPacket(pk *packet.Packet) error {
	switch flowtype.FlowType(pk.Header.FlowType) {
	default:
		return fmt.Errorf("Invalid packet type %v", pk.Header)

	case flowtype.Request:
	}
	return nil
}

func (app *App) handleRecvPacket(pk *packet.Packet) error {
	switch flowtype.FlowType(pk.Header.FlowType) {
	default:
		return fmt.Errorf("Invalid packet type %v", pk.Header)
	case flowtype.Notification:
		//process noti here
		// robj, err := unmarshalPacketFn(header, body)

	case flowtype.Response:
		// process response
		if rspfn, err := app.pid2recv.GetRspFn(pk.Header.PacketID); err != nil {
			fmt.Printf("GetRspFn %v %v\n", err, pk.Header)
			return err
		} else {
			return rspfn(pk)
		}
	}
	return nil
}

func (app *App) ReqWithRspFn(
	cmd commandid.CommandID, body interface{},
	fn packetid2rspfn.HandleRspFn) error {

	pid := app.pid2recv.NewPID(fn)
	bodybytes, err := netobj.MarshalBody_msgp(body)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	spk := &packet.Packet{
		Header: header.Header{
			CommandID: uint16(cmd),
			PacketID:  pid,
			FlowType:  uint8(flowtype.Request),
			BodyLen:   uint32(len(bodybytes)),
		},
		Body: bodybytes,
	}

	if err := app.EnqueueSendPacket(spk); err != nil {
		fmt.Printf("End %v %v %v\n", app, spk, err)
		app.sendRecvStop()
		return fmt.Errorf("Send fail %v %v", app, err)
	}
	return nil
}
