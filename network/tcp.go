package network

import (
	"fmt"
	"net"
	"shila/log"
	"shila/shila"
	"strconv"
)

var _ clientEndpoint = (*ClientEndpoint)(nil)
var _ serverEndpoint = (*ServerEndpoint)(nil)
var _ packet = (*Packet)(nil)
var _ address = (*Address)(nil)
var _ path = (*Path)(nil)

type Endpoint struct{}

type ClientEndpoint struct {
	connectedTo   address
	ingressBuffer chan *packet
	egressBuffer  chan *packet
	connection    *net.TCPConn
}

func (c ClientEndpoint) New(connectTo address, connectVia path, ingressBuf chan *packet, egressBuf chan *packet) clientEndpoint {
	_ = connectVia
	return ClientEndpoint{connectedTo: connectTo, ingressBuffer: ingressBuf, egressBuffer: egressBuf, connection: nil}
}

func (c ClientEndpoint) Setup() error {

	if c.IsSetup() {
		return Error(fmt.Sprint("Unable to setup client endpoint",
			" - ", "Already setup."))
	}

	var dest = c.connectedTo.(Address).Addr
	var err error

	// Establish a connection to the server endpoint
	if c.connection, err = net.DialTCP(dest.Network(), nil, &dest); err != nil {
		c.connection = nil
		return Error(fmt.Sprint("Unable to setup client endpoint",
			" - ", err.Error()))

	}
	log.Verbose.Println("Successful established connection to", dest.String())

	// Get reader and writer reader

	return nil
}

func (c ClientEndpoint) TearDown() error {
	panic("implement me")
}

func (c ClientEndpoint) IsSetup() bool {
	return c.connection != nil
}

type ServerEndpoint struct{}

func (s ServerEndpoint) New(listenTo address, ingressBuf chan *packet, egressBuf chan *packet) serverEndpoint {
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

func (p Packet) SetPayload(payload shila.IPv4TCPPacket) {
	panic("implement me")
}

func (p Packet) GetPayload() shila.IPv4TCPPacket {
	panic("implement me")
}

type Address struct {
	Addr net.TCPAddr
}

// First argument: <ip>
// Second argument: <port>
func (a Address) New(s ...string) error {

	var err error
	if len(s) == 2 {
		var ip net.IP
		if ip = net.ParseIP(s[0]); ip != nil {
			return Error(fmt.Sprint("Invalid IPv4TCPPacket: ", s[0]))
		}
		var port int
		if port, err = strconv.Atoi(s[1]); err != nil {
			return Error(fmt.Sprint("Invalid Port: ", s[1]))
		}
		a.Addr = net.TCPAddr{ip, port, ""}
		return nil
	}

	return Error(fmt.Sprint("Invalid number of arguments."))
}

func (a Address) String() string {
	return a.Addr.String()
}

type Path struct{}

func (p Path) New(s ...string) error {
	// No path functionality w/ plain TCP.
	_ = s
	return nil
}

func (p Path) String() string {
	// No path functionality w/ plain TCP.
	_ = p
	return ""
}
