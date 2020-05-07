package networkEndpoint

import (
	"fmt"
	"net"
	"shila/config"
	"shila/core/model"
	"shila/log"
)

var _ model.ClientNetworkEndpoint = (*Client)(nil)

type Client struct {
	connectedTo Address
	connection  *net.TCPConn
	Base
	isRunning 	bool
}

func newClient(connectTo model.NetworkAddress, connectVia model.NetworkPath,
	label model.EndpointLabel, config config.NetworkEndpoint) model.ClientNetworkEndpoint {
	_ = connectVia
	return &Client{connectTo.(Address),nil,Base{label, model.TrafficChannels{}, config},false}
}

func (c *Client) Key() model.EndpointKey {
	return model.EndpointKey(Generator{}.GetAddressPathKey(c.connectedTo, Generator{}.NewPath("")))
}

func (c *Client) SetupAndRun() error {

	if c.IsRunning() {
		return Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ",c.Key(), "}. - Client is already running."))
	}

	// Establish a connection to the server endpoint
	connection, err := net.DialTCP(c.connectedTo.Addr.Network(), nil, &c.connectedTo.Addr)
	if err != nil {
		return Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ",c.Key(), "}. - ", err.Error()))
	}
	c.connection = connection

	// Create the channels
	c.trafficChannels.Ingress = make(chan *model.Packet, c.config.SizeIngressBuff)
	c.trafficChannels.Egress  = make(chan *model.Packet, c.config.SizeEgressBuff)
	c.trafficChannels.Label   = c.Label()
	c.trafficChannels.Key	  = c.Key()

	log.Verbose.Print("Client {", c.Label(), " ", c.Key(), "} established connection to ", c.connectedTo.String(), ".")

	c.isRunning = true
	return nil
}

func (c *Client) TearDown() error {
	return nil
}

func (c *Client) IsRunning() bool {
	return c.isRunning
}

func (c *Client) TrafficChannels() model.TrafficChannels {
	return c.trafficChannels
}

func (c *Client) Label() model.EndpointLabel {
	return c.label
}

func (c *Client) handleConnection(connection *net.TCPConn) {

}

