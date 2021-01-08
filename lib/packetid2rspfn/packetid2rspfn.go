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

package packetid2rspfn

import (
	"fmt"
	"sync"

	"github.com/kasworld/mininet/lib/packet"
)

type HandleRspFn func(pk *packet.Packet) error
type PID2RspFn struct {
	mutex      sync.Mutex
	pid2recvfn map[uint32]HandleRspFn
	pid        uint32
}

func New() *PID2RspFn {
	rtn := &PID2RspFn{
		pid2recvfn: make(map[uint32]HandleRspFn),
	}
	return rtn
}
func (p2r *PID2RspFn) NewPID(fn HandleRspFn) uint32 {
	p2r.mutex.Lock()
	defer p2r.mutex.Unlock()
	p2r.pid++
	p2r.pid2recvfn[p2r.pid] = fn
	return p2r.pid
}
func (p2r *PID2RspFn) GetRspFn(pid uint32) (HandleRspFn, error) {
	p2r.mutex.Lock()
	defer p2r.mutex.Unlock()
	if recvfn, exist := p2r.pid2recvfn[pid]; exist {
		delete(p2r.pid2recvfn, pid)
		return recvfn, nil
	}
	return nil, fmt.Errorf("pid not found %v", pid)
}
