package main

import (
	"github.com/jennal/goplay-connector/connector"
	"github.com/jennal/goplay-master/master"
	"github.com/jennal/goplay/cmd"
	"github.com/jennal/goplay/log"
	"github.com/jennal/goplay/service"
	"github.com/jennal/goplay/transfer/tcp"
)

func main() {
	ser := tcp.NewServer("", connector.PORT)
	serv := service.NewService(connector.NAME, ser)

	filter, err := connector.NewFilter(ser, "", master.PORT)
	if err != nil {
		log.Error(err)
	}
	serv.RegistFilter(filter)
	cmd.Start(serv)
}
