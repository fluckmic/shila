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
}

func (c *Client) Key() model.EndpointKey {
	return model.EndpointKey(Generator{}.GetAddressPathKey(c.connectedTo, Generator{}.NewPath("")))
}

func newClient(connectTo model.NetworkAddress, connectVia model.NetworkPath,
	label model.EndpointLabel, config config.NetworkEndpoint) model.ClientNetworkEndpoint {
	_ = connectVia
	return &Client{connectTo.(Address), nil, Base{label, model.TrafficChannels{}}}
}

func (c *Client) SetupAndRun() error {

	if c.IsSetup() {
		return Error(fmt.Sprint("Unable to setup client endpoint",
			" - ", "Already setup."))
	}

	var dest = c.connectedTo.Addr
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

func (c *Client) TearDown() error {
	return nil
}

func (c *Client) IsSetup() bool {
	return c.connection != nil
}

func (c *Client) TrafficChannels() model.TrafficChannels {
	return c.trafficChannels
}

func (c *Client) Label() model.EndpointLabel {
	return c.label
}