package networkEndpoint

import (
	"fmt"
	"net"
	"shila/config"
	"shila/log"
	"shila/shila"
	"strconv"
)

var _ shila.NetworkEndpointGenerator 	= (*Generator)(nil)
var _ shila.NetworkAddressGenerator  	= (*Generator)(nil)
var _ shila.NetworkPathGenerator 	 	= (*Generator)(nil)
var _ shila.NetworkKeyGenerator			= (*Generator)(nil)
var _ shila.ClientNetworkEndpoint 		= (*Client)(nil)
var _ shila.ServerNetworkEndpoint 		= (*Server)(nil)
var _ shila.NetworkAddress 				= (*Address)(nil)
var _ shila.NetworkPath 				= (*Path)(nil)

type Generator struct{}

func (g Generator) GetDefaultContactingPath(address shila.NetworkAddress) shila.NetworkPath {
	_ = address
	return g.NewPath("")
}

func (g Generator) GetAddressKey(address shila.NetworkAddress) shila.Key_NetworkAddress_ {
	return shila.Key_NetworkAddress_(address.String())
}

func (g Generator) GetAddressPathKey(address shila.NetworkAddress, path shila.NetworkPath) shila.Key_NetworkAddressAndPath_ {
	_ = path
	return shila.Key_NetworkAddressAndPath_(address.String())
}

type Base struct {
	label shila.EndpointLabel
	trafficChannels shila.TrafficChannels
}

func (g Generator) NewClient(connectTo shila.NetworkAddress, connectVia shila.NetworkPath,
	label shila.EndpointLabel, config config.NetworkEndpoint) shila.ClientNetworkEndpoint {
	return newClient(connectTo, connectVia, label, config)
}

func (g Generator) NewServer(listenTo shila.NetworkAddress, label shila.EndpointLabel,
	config config.NetworkEndpoint) shila.ServerNetworkEndpoint {
	return newServer(listenTo, label, config)
}

func (g Generator) NewAddress(address string) shila.NetworkAddress {
	return newAddress(address)
}

func (g Generator) NewLocalAddress(port string) shila.NetworkAddress {
	return newLocalNetworkAddress(port)
}

func (g Generator) NewPath(path string) shila.NetworkPath {
	return newPath(path)
}

type Client struct {
	connectedTo Address
	connection  *net.TCPConn
	Base
}

func newClient(connectTo shila.NetworkAddress, connectVia shila.NetworkPath,
	label shila.EndpointLabel, config config.NetworkEndpoint) shila.ClientNetworkEndpoint {
	_ = connectVia
	return &Client{connectTo.(Address), nil, Base{label, shila.TrafficChannels{}}}
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

func newServer(listenTo shila.NetworkAddress, label shila.EndpointLabel, config config.NetworkEndpoint) shila.ServerNetworkEndpoint {
	return &Server{listenTo, Base{label, shila.TrafficChannels{}}}
}

func (s *Server) SetupAndRun() error {
	return nil
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

// <ip>:<port>
func newAddress(address string) shila.NetworkAddress {

	if host, port, err := net.SplitHostPort(address); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new address from ", address, "."))
		return nil
	} else {
		IPv4 := net.ParseIP(host)
		Port, err := strconv.Atoi(port)
		if IPv4 != nil || err != nil {
			log.Error.Panic(fmt.Sprint("Unable to create new address from ", address, "."))
			return nil
		} else {
			return Address{Addr: net.TCPAddr{IP: IPv4, Port: Port}}
		}
	}

	return nil
}

// <port>
func newLocalNetworkAddress(port string) shila.NetworkAddress {
	if Port, err := strconv.Atoi(port); err != nil {
		log.Error.Panic(fmt.Sprint("Unable to create new local address from ", port, "."))
		return nil
	} else {
		return Address{Addr: net.TCPAddr{Port: Port}}
	}
return nil
}

func (a Address) String() string {
	return a.Addr.String()
}

type Path struct{}

func newPath(path string) shila.NetworkPath {
	// No path functionality w/ plain TCP.
	_ = path; return Path{}
}

func (p Path) String() string {
	// No path functionality w/ plain TCP.
	return ""
}
