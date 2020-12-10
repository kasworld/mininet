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

package flowtype

import "fmt"

// FlowType packet flow type
type FlowType byte

const (
	// make uninitalized packet error
	invalid FlowType = iota
	// Request for request packet (response packet expected)
	Request
	// Response is reply of request packet
	Response
	// Notification is just send and forget packet
	Notification
)

var _flowTypeStr = map[FlowType]string{
	invalid:      "invalid",
	Request:      "Request",
	Response:     "Response",
	Notification: "Notification",
}

func (e FlowType) String() string {
	if s, exist := _flowTypeStr[e]; exist {
		return s
	}
	return fmt.Sprintf("FlowType%d", byte(e))
}
