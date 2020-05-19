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
	Base
	connection			networkConnection
	ingressRaw 			chan byte
	isValid    			bool
	isSetup    			bool // TODO: merge to "state" object
	isRunning 			bool
}

type networkConnection struct {
	Identifier model.NetworkConnectionIdentifier
	Backbone   *net.TCPConn
}

func newClient(netConnId model.NetworkConnectionIdentifier, label model.EndpointLabel, config config.NetworkEndpoint) model.ClientNetworkEndpoint {
	return &Client{
		Base: 				Base{
								label: label,
								config: config,
							},
		connection:		    networkConnection{Identifier: netConnId},
		ingressRaw: 		make(chan byte, config.SizeReadBuffer),
		isValid: 			true,
	}
}

func (c *Client) Key() model.EndpointKey {
	return model.EndpointKey(model.KeyGenerator{}.NetworkAddressAndPathKey(c.connection.Identifier.Dst, Generator{}.NewPath("")))
}

func (c *Client) SetupAndRun() (model.NetworkConnectionIdentifier, error) {

	if !c.IsValid() {
		return model.NetworkConnectionIdentifier{}, Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(), "}. - Client no longer valid."))
	}

	if c.IsRunning() {
		return model.NetworkConnectionIdentifier{}, Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(), "}. - Client is already running."))
	}

	if c.IsSetup() {
		return model.NetworkConnectionIdentifier{}, nil
	}

	// Establish a connection to the server endpoint
	dst := c.connection.Identifier.Dst.(Address)
	backboneConnection, err := net.DialTCP(dst.Addr.Network(), nil, &dst.Addr)
	if err != nil {
		return model.NetworkConnectionIdentifier{},
		Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(),
		"}. - Unable to setup the backbone connection. - ", err.Error()))
	}
	c.connection.Backbone = backboneConnection

	if c.Label() == model.TrafficNetworkEndpoint {
		// Before setting the own src address, a traffic client sends the currently set src address to the server;
		// which should be (or is.) the src address of the corresponding contacting client endpoint. This information
		// is required to be able to do the mapping on the server side.

		//msg := fmt.Sprint(c.connection.Identifier.Src.String(),'\n')
		if _, err := c.connection.Backbone.Write([]byte("Hello\n")); err != nil {
			return model.NetworkConnectionIdentifier{},
				Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(),
					"}. - Unable to send source address. - ", err.Error()))
		}
	}

	c.connection.Identifier.Src = Address{Addr: *backboneConnection.LocalAddr().(*net.TCPAddr)}

	// Create the channels
	c.ingress = make(chan *model.Packet, c.config.SizeIngressBuff)
	c.egress  = make(chan *model.Packet, c.config.SizeEgressBuff)

	go c.packetize()
	go c.serveIngress()
	go c.serveEgress()

	log.Verbose.Print("Client {", c.Label()," ", c.Key(),
	"} successfully established connection to {", c.connection.Identifier.Dst, "}.")

	c.isSetup   = true
	c.isRunning = true

	return c.connection.Identifier, nil
}

func (c *Client) TearDown() error {

	log.Verbose.Print("Tear down client {", c.Label()," ", c.Key(),
	"}} connecting to {", c.connection.Identifier.Dst, "}.")

	c.isValid = false
	c.isRunning = false
	c.isSetup = false

	err := c.connection.Backbone.Close()

	return err
}

func (c *Client) TrafficChannels() model.PacketChannels {
	return model.PacketChannels{Ingress: c.ingress, Egress: c.egress}
}

func (c *Client) Label() model.EndpointLabel {
	return c.label
}

func (c *Client) serveIngress() {
	reader := io.Reader(c.connection.Backbone)
	storage := make([]byte, c.config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, c.config.BatchSizeRead)
		log.Verbose.Print("Client {", c.Label()," ", c.Key(), "} read {",nBytesRead,"} bytes from input.")
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
	writer := io.Writer(c.connection.Backbone)
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
			panic(fmt.Sprint("Unable to get IP networkConnectionId in packetizer of client {", c.Key(),
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
