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

type Base struct {
	label shila.EndpointLabel
	trafficChannels shila.TrafficChannels
}

func (g Generator) NewClient(connectTo shila.NetworkAddress, connectVia shila.NetworkPath,
	label shila.EndpointLabel, channels shila.TrafficChannels) shila.ClientNetworkEndpoint {
	return newClient(connectTo, connectVia, label, channels)
}

func (g Generator) NewServer(listenTo shila.NetworkAddress, label shila.EndpointLabel,
	channels shila.TrafficChannels) shila.ServerNetworkEndpoint {
	return newServer(listenTo, label, channels)
}

type Client struct {
	connectedTo Address
	connection  *net.TCPConn
	Base
}

func newClient(connectTo shila.NetworkAddress, connectVia shila.NetworkPath,
	label shila.EndpointLabel, channels shila.TrafficChannels) shila.ClientNetworkEndpoint {
	_ = connectVia
	return &Client{connectTo.(Address), nil, Base{label, channels}}
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

func (c *Client) TrafficChannels() shila.TrafficChannels {
	return c.trafficChannels
}

func (c *Client) Label() shila.EndpointLabel {
	return c.label
}

type Server struct{
	listenTo shila.NetworkAddress
	Base
}

func newServer(listenTo shila.NetworkAddress, label shila.EndpointLabel, channels shila.TrafficChannels) shila.ServerNetworkEndpoint {
	return &Server{listenTo, Base{label, channels}}
}

func (s *Server) SetupAndRun() error {
	panic("implement me")
}

func (s *Server) TearDown() error {
	panic("implement me")
}

func (s *Server) TrafficChannels() shila.TrafficChannels {
	return s.trafficChannels
}

func (s *Server) Label() shila.EndpointLabel {
	return s.label
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
