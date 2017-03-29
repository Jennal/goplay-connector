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
	serviceInfos map[string]master.ServicePack

	servicesMutex sync.Mutex
	services      map[string]*service.ServiceClient

	masterClient *master.MasterClient

	infoMutex sync.Mutex
	info      master.ServicePack
}

func NewFilter(server transfer.IServer, host string, port int) (*Filter, error) {
	ins := &Filter{
		server:       server,
		serviceInfos: make(map[string]master.ServicePack),
		services:     make(map[string]*service.ServiceClient),
		masterClient: master.NewMasterClient(tcp.NewClient()),
		info:         master.NewServicePack(master.ST_CONNECTOR, NAME, server.Port()),
	}

	err := ins.masterClient.Bind(server, &ins.info, host, port)
	if err != nil {
		return nil, err
	}

	return ins, nil
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
	if s, ok := self.services[name]; ok {
		self.servicesMutex.Unlock()
		if s.IsConnected() {
			return s
		}
	}
	self.servicesMutex.Unlock()

	sp, ok := self.serviceInfos[name]
	if !ok {
		psp, err := self.masterClient.GetByName(name)
		if err != nil {
			log.Error(err)
			return nil
		}

		self.serviceInfos[name] = psp
	}

	sp, ok = self.serviceInfos[name]
	if !ok {
		log.Errorf("get service info failed!")
		return nil
	}

	cli := tcp.NewClient()
	client := service.NewServiceClient(cli)
	client.RegistFilter(NewRpcFilter(self.server, client))

	client.On(transfer.EVENT_CLIENT_CONNECTED, self, func(cli transfer.IClient) {
		self.servicesMutex.Lock()
		defer self.servicesMutex.Unlock()
		self.services[name] = client
	})
	client.On(transfer.EVENT_CLIENT_DISCONNECTED, self, func(cli transfer.IClient) {
		self.servicesMutex.Lock()
		defer self.servicesMutex.Unlock()

		delete(self.services, name)
	})

	err := client.Connect(sp.IP, sp.Port)
	if err != nil {
		log.Error(err)
		return nil
	}

	return client
}

func (self *Filter) OnNewClient(sess *session.Session) bool /* return false to ignore */ {
	//TODO: check if ip block

	return true
}

func (self *Filter) OnRecv(sess *session.Session, header *pkg.Header, body []byte) bool /* return false to ignore */ {
	if header.Type == pkg.PKG_HEARTBEAT || header.Type == pkg.PKG_HEARTBEAT_RESPONSE {
		return true
	}

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
