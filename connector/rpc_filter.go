// Copyright (C) 2017 Jennal(jennalcn@gmail.com). All rights reserved.
//
// Licensed under the MIT License (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software distributed 
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR 
// CONDITIONS OF ANY KIND, either express or implied. See the License for the 
// specific language governing permissions and limitations under the License.

package connector

import (
	"github.com/jennal/goplay/pkg"
	"github.com/jennal/goplay/service"
	"github.com/jennal/goplay/session"
	"github.com/jennal/goplay/transfer"
)

type RpcFilter struct {
	server transfer.IServer
	client *service.ServiceClient
}

func NewRpcFilter(server transfer.IServer, client *service.ServiceClient) *RpcFilter {
	return &RpcFilter{
		server: server,
		client: client,
	}
}

func (self *RpcFilter) OnNewClient(sess *session.Session) bool /* return false to ignore */ {
	return true
}

func (self *RpcFilter) OnRecv(sess *session.Session, header *pkg.Header, body []byte) bool /* return false to ignore */ {
	if header.Type == pkg.PKG_HEARTBEAT || header.Type == pkg.PKG_HEARTBEAT_RESPONSE {
		return true
	}

	cli := self.server.GetClientById(header.ClientID)
	if cli != nil {
		h := pkg.NewHeaderFromRpc(header)
		cli.Send(h, body)
	}

	return false
}
