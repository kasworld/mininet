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

package serveconn

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/kasworld/massecho/protocol_me/me_authorize"
	"github.com/kasworld/massecho/protocol_me/me_const"
	"github.com/kasworld/massecho/protocol_me/me_idcmd"
	"github.com/kasworld/massecho/protocol_me/me_idnoti"
	"github.com/kasworld/massecho/protocol_me/me_looptcp"
	"github.com/kasworld/massecho/protocol_me/me_loopwsgorilla"
	"github.com/kasworld/massecho/protocol_me/me_packet"
	"github.com/kasworld/massecho/protocol_me/me_statapierror"
	"github.com/kasworld/massecho/protocol_me/me_statnoti"
	"github.com/kasworld/massecho/protocol_me/me_statserveapi"
	"golang.org/x/net/websocket"
)

type CounterI interface {
