package tcp

import (
	"fmt"
	"net"
	"shila/log"
	"shila/shila"
	"strconv"
)

var _ shila.ClientNetworkEndpoint = (*ClientEndpoint)(nil)
var _ shila.ServerNetworkEndpoint = (*ServerEndpoint)(nil)
var _ shila.NetworkAddress = (*Address)(nil)
var _ shila.NetworkPath = (*Path)(nil)

type Endpoint struct{}

type ClientEndpoint struct {
	connectedTo   Address
	ingressBuffer shila.PacketChannel
	egressBuffer  shila.PacketChannel
	connection    *net.TCPConn
}

func (c *ClientEndpoint) New(connectTo shila.NetworkAddress, connectVia shila.NetworkPath,
	ingressBuf shila.PacketChannel, egressBuf shila.PacketChannel) shila.ClientNetworkEndpoint {
	_ = connectVia
	return &ClientEndpoint{connectedTo: connectTo.(Address), ingressBuffer: ingressBuf, egressBuffer: egressBuf, connection: nil}
}

func (c *ClientEndpoint) Setup() error {

	if c.IsSetup() {
		return shila.Error(fmt.Sprint("Unable to setup client endpoint",
			" - ", "Already setup."))
	}

	var dest = c.connectedTo.Addr
	var err error

	// Establish a connection to the server endpoint
	if c.connection, err = net.DialTCP(dest.Network(), nil, &dest); err != nil {
		c.connection = nil
		return shila.Error(fmt.Sprint("Unable to setup client endpoint",
			" - ", err.Error()))

	}
	log.Verbose.Println("Successful established connection to", dest.String())

	// Get reader and writer reader

	return nil
}

func (c *ClientEndpoint) TearDown() error {
	panic("implement me")
}

func (c *ClientEndpoint) IsSetup() bool {
	return c.connection != nil
}

func (c *ClientEndpoint) Label() shila.EndpointLabel {
	return shila.NetworkClientEndpoint
}

type ServerEndpoint struct{}

func (s *ServerEndpoint) New(listenTo shila.NetworkAddress, ingressBuf shila.PacketChannel,
	egressBuf shila.PacketChannel) shila.ServerNetworkEndpoint {
	panic("implement me")
}

func (s *ServerEndpoint) Setup() error {
	panic("implement me")
}

func (s *ServerEndpoint) TearDown() error {
	panic("implement me")
}

func (c *ServerEndpoint) Label() shila.EndpointLabel {
	return shila.NetworkServerEndpoint
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
			return shila.Error(fmt.Sprint("Invalid IPv4TCPPacket: ", s[0]))
		}
		var port int
		if port, err = strconv.Atoi(s[1]); err != nil {
			return shila.Error(fmt.Sprint("Invalid Port: ", s[1]))
		}
		a.Addr = net.TCPAddr{ip, port, ""}
		return nil
	}

	return shila.Error(fmt.Sprint("Invalid number of arguments."))
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
