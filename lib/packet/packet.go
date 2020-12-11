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

package packet

import (
	"fmt"
	"io"

	"github.com/kasworld/mininet/lib/header"
)

type Packet struct {
	Header header.Header
	Body   []byte
}

// Bytes2Packet return packet from bytelist
func Bytes2Packet(rdata []byte) (*Packet, error) {
	if len(rdata) < header.HeaderLen {
		return &Packet{}, fmt.Errorf("header not complete")
	}
	hd := header.MakeHeaderFromBytes(rdata)
	if len(rdata) != header.HeaderLen+int(hd.BodyLen) {
		return &Packet{hd, nil}, fmt.Errorf("packet not complete")
	}
	return &Packet{
		hd,
		rdata[header.HeaderLen : header.HeaderLen+int(hd.BodyLen)],
	}, nil
}

// ReadPacket return packet from io.Reader
func ReadPacket(conn io.Reader) (*Packet, error) {
	recvLen := 0
	toRead := header.HeaderLen
	readBuffer := make([]byte, toRead)
	for recvLen < toRead {
		n, err := conn.Read(readBuffer[recvLen:toRead])
		if err != nil {
			return &Packet{}, err
		}
		recvLen += n
	}
	hd := header.MakeHeaderFromBytes(readBuffer)
	recvLen = 0
	toRead = int(hd.BodyLen)
	readBuffer = make([]byte, toRead)
	for recvLen < toRead {
		n, err := conn.Read(readBuffer[recvLen:toRead])
		if err != nil {
			return &Packet{hd, nil}, err
		}
		recvLen += n
	}
	return &Packet{hd, readBuffer}, nil
}

// WritePacket write packet to io.Writer
func WritePacket(conn io.Writer, pk *Packet) error {
	sendbuf := pk.Header.ToByteList()
	toWrite := len(sendbuf)
	for l := 0; l < toWrite; {
		n, err := conn.Write(sendbuf[l:toWrite])
		if err != nil {
			return err
		}
		l += n
	}
	sendbuf = pk.Body
	toWrite = len(sendbuf)
	for l := 0; l < toWrite; {
		n, err := conn.Write(sendbuf[l:toWrite])
		if err != nil {
			return err
		}
		l += n
	}
	return nil
}
