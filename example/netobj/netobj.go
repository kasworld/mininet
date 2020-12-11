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

// Invalid not used, make empty packet error
type ReqInvalid_data struct {
	Dummy uint8 // change as you need
}

// Invalid not used, make empty packet error
type RspInvalid_data struct {
	Dummy uint8 // change as you need
}

// Echo simple echo
type ReqEcho_data struct {
	Msg string
}

// Echo simple echo
type RspEcho_data struct {
	Msg string
}

// Invalid not used, make empty packet error
type NotiInvalid_data struct {
	Dummy uint8 // change as you need
}