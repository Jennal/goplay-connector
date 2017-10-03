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
	"sync"

	"strings"

	"github.com/jennal/goplay-master/master"
	"github.com/jennal/goplay/log"
	"github.com/jennal/goplay/pkg"
	"github.com/jennal/goplay/service"
	"github.com/jennal/goplay/session"
	"github.com/jennal/goplay/transfer"
	"github.com/jennal/goplay/transfer/tcp"
)

type Filter struct {
	server       transfer.IServer
	serviceInfos map[string]*master.ServicePack

	servicesMutex sync.Mutex
	services      map[string]*service.ServiceClient

	masterClient *master.MasterClient

	infoMutex sync.Mutex
	info      *master.ServicePack
}

func NewFilter(server transfer.IServer, host string, port int) (*Filter, error) {
	ins := &Filter{
		server:       server,
		serviceInfos: make(map[string]*master.ServicePack),
		services:     make(map[string]*service.ServiceClient),
		masterClient: master.NewMasterClient(tcp.NewClient()),
		info:         master.NewServicePack(master.ST_CONNECTOR, NAME, server.Port()),
	}

	err := ins.masterClient.Bind(server, ins.info, host, port)
	if err != nil {
		return nil, err
	}

	ins.masterClient.On(master.ON_BACKEND_UPDATED, ins, func(sp *master.ServicePack) {
		if item, ok := ins.serviceInfos[sp.Name]; ok {
			if item.IP == sp.IP && item.Port == sp.Port {
				ins.servicesMutex.Lock()
				ins.serviceInfos[sp.Name] = sp
				ins.servicesMutex.Unlock()
			}

			return
		}

		ins.connectBackend(sp)
	})

	backends, e := ins.masterClient.GetBackends()
	if e == nil {
		for _, backend := range backends {
			ins.connectBackend(backend)
		}
	}

	return ins, nil
}

func (self *Filter) connectBackend(sp *master.ServicePack) {
	//check if back exists
	self.servicesMutex.Lock()
	if si, ok := self.serviceInfos[sp.Name]; ok {
		if s, ok := self.services[sp.Name]; ok &&
			si.Type == sp.Type &&
			si.IP == sp.IP &&
			si.Port == sp.Port &&
			s.IsConnected() {
			self.servicesMutex.Unlock()
			return
		}
	}
	self.servicesMutex.Unlock()

	cli := tcp.NewClient()
	client := service.NewServiceClient(cli)
	client.RegistFilter(NewRpcFilter(self.server, client))

	client.Once(transfer.EVENT_CLIENT_CONNECTED, self, func(cli transfer.IClient) {
		self.servicesMutex.Lock()
		defer self.servicesMutex.Unlock()

		// log.Log("+++++++++++++++++++++++++++++")
		self.serviceInfos[sp.Name] = sp
		self.services[sp.Name] = client
	})
	// log.Logf("************ Regist: %p | %p", client, self)
	client.Once(transfer.EVENT_CLIENT_DISCONNECTED, self, func(cli transfer.IClient) {
		self.servicesMutex.Lock()
		defer self.servicesMutex.Unlock()

		// log.Log("-----------------------------")
		delete(self.services, sp.Name)
		delete(self.serviceInfos, sp.Name)
	})

	// log.Logf("************ Connect: %#v", sp)
	err := client.Connect(sp.IP, sp.Port)
	if err != nil {
		log.Error(err)
	}
	// log.Logf("************ Connected")
}

func (self *Filter) GetServiceName(route string) string {
	arr := strings.Split(route, ".")
	if len(arr) <= 0 {
		return ""
	}

	return arr[0]
}

func (self *Filter) GetService(name string) transfer.IClient {
	self.servicesMutex.Lock()
	defer self.servicesMutex.Unlock()

	if s, ok := self.services[name]; ok {
		if s.IsConnected() {
			return s
		}
	}

	return nil
}

func (self *Filter) GetAllBackends() map[string]*service.ServiceClient {
	self.servicesMutex.Lock()
	defer self.servicesMutex.Unlock()
	backends := make(map[string]*service.ServiceClient)
	for name, item := range self.services {
		backends[name] = item
	}

	return backends
}

func (self *Filter) OnNewClient(sess *session.Session) bool /* return false to ignore */ {
	//TODO: check if ip block

	log.Logf("OnNewClient: %d => %p", sess.Id(), sess)
	/* Notify New Client to Backend */
	sess.Once(transfer.EVENT_CLIENT_DISCONNECTED, self, func(client transfer.IClient) {
		header := sess.NewHeader(pkg.PKG_RPC_NOTIFY, sess.Encoding, master.ON_CONNECTOR_CLIENT_DISCONNECTED)
		header = pkg.NewRpcHeader(header, sess.Id())

		backends := self.GetAllBackends()
		for _, backend := range backends {
			backend.Send(header, []byte{})
		}
	})

	header := sess.NewHeader(pkg.PKG_RPC_NOTIFY, sess.Encoding, master.ON_CONNECTOR_GOT_NET_CLIENT)
	header = pkg.NewRpcHeader(header, sess.Id())

	backends := self.GetAllBackends()
	for _, backend := range backends {
		backend.Send(header, []byte{})
	}

	return true
}

func (self *Filter) OnRecv(sess *session.Session, header *pkg.Header, body []byte) bool /* return false to ignore */ {
	if header.Type == pkg.PKG_HEARTBEAT || header.Type == pkg.PKG_HEARTBEAT_RESPONSE {
		return true
	}

	log.Logf("OnRecv: %d | %p => %#v", sess.Id(), sess, sess)
	name := self.GetServiceName(header.Route)
	if name == "" {
		return true
	}

	s := self.GetService(name)
	if s == nil {
		return true
	}

	h := pkg.NewRpcHeader(header, sess.Id())
	s.Send(h, body)
	return false
}
