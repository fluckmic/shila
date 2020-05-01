package networkEndpoint

import (
	"fmt"
	"net"
	"shila/log"
	"shila/shila"
	"strconv"
)

var _ shila.NetworkEndpoint = (*Generator)(nil)
var _ shila.ClientNetworkEndpoint = (*Client)(nil)
var _ shila.ServerNetworkEndpoint = (*Server)(nil)
var _ shila.NetworkAddress = (*Address)(nil)
var _ shila.NetworkPath = (*Path)(nil)

type Generator struct{}

func (g Generator) NewClient(cT shila.NetworkAddress, cV shila.NetworkPath,
	iB shila.PacketChannel, eB shila.PacketChannel) shila.ClientNetworkEndpoint {
	return newClient(cT, cV, iB, eB)
}

func (g Generator) NewServer(lT shila.NetworkAddress, iB shila.PacketChannel,
	eB shila.PacketChannel) shila.ServerNetworkEndpoint {
	return newServer(lT, iB, eB)
}

type Client struct {
	connectedTo   Address
	ingressBuffer shila.PacketChannel
	egressBuffer  shila.PacketChannel
	connection    *net.TCPConn
}

func newClient(connectTo shila.NetworkAddress, connectVia shila.NetworkPath,
	ingressBuf shila.PacketChannel, egressBuf shila.PacketChannel) shila.ClientNetworkEndpoint {
	_ = connectVia
	return &Client{connectedTo: connectTo.(Address), ingressBuffer: ingressBuf, egressBuffer: egressBuf, connection: nil}
}

func (c *Client) SetupAndRun() error {

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

func (c *Client) TearDown() error {
	panic("implement me")
}

func (c *Client) IsSetup() bool {
	return c.connection != nil
}

func (c *Client) Label() shila.EndpointLabel {
	return shila.NetworkClientEndpoint
}

type Server struct{}

func newServer(listenTo shila.NetworkAddress, ingressBuf shila.PacketChannel,
	egressBuf shila.PacketChannel) shila.ServerNetworkEndpoint {
	panic("implement me")
}

func (s *Server) SetupAndRun() error {
	panic("implement me")
}

func (s *Server) TearDown() error {
	panic("implement me")
}

func (c *Server) Label() shila.EndpointLabel {
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
