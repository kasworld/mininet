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

package netobj

import (
	"bytes"
	"encoding/gob"

	"github.com/kasworld/mininet/lib/packet"
	"github.com/tinylib/msgp/msgp"
)

func MarshalBody_gob(body interface{}) ([]byte, error) {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	err := enc.Encode(body)
	return network.Bytes(), err
}

func MarshalBody_msgp(body interface{}) ([]byte, error) {
	return body.(msgp.Marshaler).MarshalMsg(nil)
}

func Unmarshal_ReqEcho_gob(pk *packet.Packet) (interface{}, error) {
	var args ReqEcho_data
	network := bytes.NewBuffer(pk.Body)
	dec := gob.NewDecoder(network)
	err := dec.Decode(&args)
	return &args, err
}

func Unmarshal_ReqEcho_msgp(pk *packet.Packet) (interface{}, error) {
	var args ReqEcho_data
	if _, err := args.UnmarshalMsg(pk.Body); err != nil {
		return nil, err
	}
	return &args, nil
}
