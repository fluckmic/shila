package network

import (
	"fmt"
	"net"
	"shila/shila"
	"strconv"
)

var _ clientEndpoint = (*ClientEndpoint)(nil)
var _ serverEndpoint = (*ServerEndpoint)(nil)
var _ packet = (*Packet)(nil)
var _ address = (*Address)(nil)
var _ path = (*Path)(nil)

type Endpoint struct{}

type ClientEndpoint struct{}

func (c ClientEndpoint) New(connectTo address, connectVia path, ingressBuf chan *packet, egressBuf chan *packet) {
	panic("implement me")
}

func (c ClientEndpoint) Setup() error {
	panic("implement me")
}

func (c ClientEndpoint) TearDown() error {
	panic("implement me")
}

type ServerEndpoint struct{}

func (s ServerEndpoint) New(listenTo address, ingressBuf chan *packet, egressBuf chan *packet) {
	panic("implement me")
}

func (s ServerEndpoint) Setup() error {
	panic("implement me")
}

func (s ServerEndpoint) TearDown() error {
	panic("implement me")
}

type Packet struct{}

func (p Packet) SetAddress(address address) {
	panic("implement me")
}

func (p Packet) GetAddress() address {
	panic("implement me")
}

func (p Packet) SetPayload(payload shila.IPPacket) {
	panic("implement me")
}

func (p Packet) GetPayload() shila.IPPacket {
	panic("implement me")
}

type Address struct {
	addr net.TCPAddr
}

// First argument: <ip>
// Second argument: <port>
func (a Address) New(s ...string) error {

	var err error
	if len(s) == 2 {
		var ip net.IP
		if ip = net.ParseIP(s[0]); ip != nil {
			return Error(fmt.Sprint("Invalid IP: ", s[0]))
		}
		var port int
		if port, err = strconv.Atoi(s[1]); err != nil {
			return Error(fmt.Sprint("Invalid Port: ", s[1]))
		}
		a.addr = net.TCPAddr{ip, port, ""}
		return nil
	}

	return Error(fmt.Sprint("Invalid number of arguments."))
}

func (a Address) String() string {
	return a.addr.String()
}

type Path struct{}

func (p Path) New(s ...string) error {
	panic("implement me")
}

func (p Path) String() string {
	panic("implement me")
}
