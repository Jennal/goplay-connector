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

package main

import (
	"flag"

	"github.com/jennal/goplay-connector/connector"
	"github.com/jennal/goplay-master/master"
	"github.com/jennal/goplay/cmd"
	"github.com/jennal/goplay/log"
	"github.com/jennal/goplay/pkg"
	"github.com/jennal/goplay/service"
	"github.com/jennal/goplay/transfer/tcp"
)

var (
	port        *int
	master_host *string
	master_port *int
)

func init() {
	port = flag.Int("p", connector.PORT, "port of the server")
	master_host = flag.String("master-host", "", "host of master server")
	master_port = flag.Int("master-port", master.PORT, "port of master server")
}

func main() {
	flag.Parse()
	pkg.SetHandShakeImpl(connector.NewHandShake())

	ser := tcp.NewServer("", *port)
	serv := service.NewService(connector.NAME, ser)

	filter, err := connector.NewFilter(ser, *master_host, *master_port)
	if err != nil {
		log.Error(err)
		return
	}

	serv.RegistFilter(filter)
	cmd.Start(serv)
}
