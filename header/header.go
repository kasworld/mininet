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

package header

import (
	"encoding/binary"
	"fmt"
)

const (
	// HeaderLen fixed size of header
	HeaderLen = 4 + 4 + 2 + 2 + 1 + 1 + 2

	// MaxBodyLen max body size byte of raw packet
	MaxBodyLen = 0xfffff

	// MaxPacketLen max total packet size byte of raw packet
	MaxPacketLen = HeaderLen + MaxBodyLen
)

// Header is fixed size header of packet
type Header struct {
	BodyLen     uint32 // set after marshal body
	PacketID    uint32 // unique id per packet (wrap around reuse)
	CommandID   uint16 // application demux received packet
	ResultCode  uint16 // for Response
	FlowType    byte   // flow control, Request, Response, Notification
	MarshalType byte   // body marshal,compress type
	ExtraData   uint16 // any data
}

func (h Header) String() string {
	return fmt.Sprintf(
		"Header[%v:%v PacketID:%v ResultCode:%v BodyLen:%v MarshalType:%v ExtraData:%v]",
		h.FlowType, h.CommandID, h.PacketID, h.ResultCode, h.BodyLen, h.MarshalType, h.ExtraData)
}

// MakeHeaderFromBytes unmarshal header from bytelist
func MakeHeaderFromBytes(buf []byte) Header {
	var h Header
	h.BodyLen = binary.LittleEndian.Uint32(buf[0:4])
	h.PacketID = binary.LittleEndian.Uint32(buf[4:8])
	h.CommandID = binary.LittleEndian.Uint16(buf[8:10])
	h.ResultCode = binary.LittleEndian.Uint16(buf[10:12])
	h.FlowType = byte(buf[12])
	h.MarshalType = buf[13]
	h.ExtraData = binary.LittleEndian.Uint16(buf[14:16])
	return h
}

func (h Header) toBytesAt(buf []byte) {
	binary.LittleEndian.PutUint32(buf[0:4], h.BodyLen)
	binary.LittleEndian.PutUint32(buf[4:8], h.PacketID)
	binary.LittleEndian.PutUint16(buf[8:10], h.CommandID)
	binary.LittleEndian.PutUint16(buf[10:12], uint16(h.ResultCode))
	buf[12] = byte(h.FlowType)
	buf[13] = h.MarshalType
	binary.LittleEndian.PutUint16(buf[14:16], h.ExtraData)
}

// ToByteList marshal header to bytelist
func (h Header) ToByteList() []byte {
	buf := make([]byte, HeaderLen)
	h.toBytesAt(buf)
	return buf
}
