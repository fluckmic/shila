package networkEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/config"
	"shila/core/model"
	"shila/layer"
	"shila/log"
)

var _ model.ClientNetworkEndpoint = (*Client)(nil)

type Client struct {
	connectedTo Address
	connection  *net.TCPConn
	Base
	ingressRaw chan byte
	isValid    bool
	isSetup    bool // TODO: merge to "state" object
	isRunning  bool
}

func newClient(connectTo model.NetworkAddress, connectVia model.NetworkPath,
	label model.EndpointLabel, config config.NetworkEndpoint) model.ClientNetworkEndpoint {
	_ = connectVia
	return &Client{
		connectedTo: connectTo.(Address),
		Base: Base{
			label: label,
			config: config,
		},
		ingressRaw: make(chan byte, config.SizeReadBuffer),
		isValid: true,
	}
}

func (c *Client) Key() model.EndpointKey {
	return model.EndpointKey(Generator{}.GetAddressPathKey(c.connectedTo, Generator{}.NewPath("")))
}

func (c *Client) SetupAndRun() error {

	if !c.IsValid() {
		return Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ",c.Key(), "}. - Client no longer valid."))
	}

	if c.IsRunning() {
		return Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ",c.Key(), "}. - Client is already running."))
	}

	if c.IsSetup() {
		return nil
	}

	// Establish a connection to the server endpoint
	connection, err := net.DialTCP(c.connectedTo.Addr.Network(), nil, &c.connectedTo.Addr)
	if err != nil {
		return Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ",c.Key(), "}. - ", err.Error()))
	}
	c.connection = connection

	// Create the channels
	c.ingress = make(chan *model.Packet, c.config.SizeIngressBuff)
	c.egress  = make(chan *model.Packet, c.config.SizeEgressBuff)

	go c.packetize()
	go c.serveIngress()
	go c.serveEgress()

	log.Verbose.Print("Client {",c.Label(),"} successfully established connection to {",c.connectedTo.String(),"}.")

	c.isSetup   = true
	c.isRunning = true
	return nil
}

func (c *Client) TearDown() error {

	log.Verbose.Print("Tear down client {",c.Label(),"} connecting to {",c.connectedTo.String(),"}.")

	c.isValid = false
	c.isRunning = false
	c.isSetup = false

	err := c.connection.Close()

	return err
}

func (c *Client) TrafficChannels() model.PacketChannels {
	return model.PacketChannels{Ingress: c.ingress, Egress: c.egress}
}

func (c *Client) Label() model.EndpointLabel {
	return c.label
}

func (c *Client) serveIngress() {
	reader := io.Reader(c.connection)
	storage := make([]byte, c.config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, c.config.BatchSizeRead)
		log.Verbose.Print("Client  {", c.Label()," ", c.Key(), "} read {",nBytesRead,"} bytes from input.")
		if err != nil && !c.IsValid() {
			// Client is no longer valid, there is no need to try to stay alive.
			return
		} else if err != nil {
			// Client is still valid, that is, a connection relies on this client.
			// Client should try to recover somehow to reestablish a connection.
			panic(fmt.Sprint("Client {", c.Label()," ", c.Key(), "} unable to read data from underlying connection. - ",
				err.Error())) // TODO: Handle panic!
		}
		for _, b := range storage[:nBytesRead] {
			c.ingressRaw <- b
		}
	}
}

func (c *Client) serveEgress() {
	writer := io.Writer(c.connection)
	for p := range c.egress {
		_, err := writer.Write(p.GetRawPayload())
		if err != nil && !c.IsValid() {
			// Error doesn't matter, client is no longer valid anyway.
			return
		} else if err != nil {
			// Client is still valid, that is, a connection relies on this client.
			// Client should try to recover somehow to reestablish a connection.
			panic(fmt.Sprint("Client {", c.Label()," ",c.Key(), "} unable to write data to underlying connection. - ",
				err.Error())) // TODO: Handle panic!
		}
	}
}

func (c *Client) packetize() {
	for {
		rawData  := layer.PacketizeRawData(c.ingressRaw, c.config.SizeReadBuffer)
		if iPHeader, err := layer.GetIPHeader(rawData); err != nil {
			panic(fmt.Sprint("Unable to get IP header in packetizer of client {", c.Key(),
				"}. - ", err.Error())) // TODO: Handle panic!
		} else {
			c.ingress <- model.NewPacket(c, iPHeader, rawData)
		}
	}
}

func (c *Client) IsValid() bool {
	return c.isValid
}

func (c *Client) IsSetup() bool {
	return c.isSetup
}

func (c *Client) IsRunning() bool {
	return c.isRunning
}
