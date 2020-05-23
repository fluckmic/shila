package networkEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/layer/tcpip"
	"shila/log"
	"shila/networkSide/network"
	"time"
)

var _ shila.ClientNetworkEndpoint = (*Client)(nil)

type Client struct {
	Base
	connection			networkConnection
	isValid    			bool
	isSetup    			bool // TODO: merge to "state" object
	isRunning 			bool
}

type networkConnection struct {
	Identifier shila.NetFlow
	Backbone   *net.TCPConn
}

func NewClient(netConnId shila.NetFlow, label shila.EndpointLabel, config config.NetworkEndpoint) shila.ClientNetworkEndpoint {
	return &Client{
		Base: 				Base{
								label: label,
								config: config,
							},
		connection:		    networkConnection{Identifier: netConnId},
		isValid: 			true,
	}
}

func (c *Client) Key() shila.EndpointKey {
	return shila.EndpointKey(shila.GetNetworkAddressAndPathKey(c.connection.Identifier.Dst, network.PathGenerator{}.New("")))
}

func (c *Client) SetupAndRun() (shila.NetFlow, error) {

	if !c.IsValid() {
		return shila.NetFlow{}, Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(), "}. - Client no longer valid."))
	}

	if c.IsRunning() {
		return shila.NetFlow{}, Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(), "}. - Client is already running."))
	}

	if c.IsSetup() {
		return shila.NetFlow{}, nil
	}

	// Establish a connection to the server endpoint
	dst := c.connection.Identifier.Dst.(*net.TCPAddr)
	backboneConnection, err := net.DialTCP(dst.Network(), nil, dst)
	if err != nil {
		return shila.NetFlow{},
		Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(),
		"}. - Unable to setup the backbone connection. - ", err.Error()))
	}
	c.connection.Backbone = backboneConnection

	if c.Label() == shila.TrafficNetworkEndpoint {
		// Before setting the own src address, a traffic client sends the currently set src address to the server;
		// which should be (or is.) the src address of the corresponding contacting client endpoint. This information
		// is required to be able to do the mapping on the server side.

		if _, err := c.connection.Backbone.Write([]byte(fmt.Sprintln(c.connection.Identifier.Src.String()))); err != nil {
			return shila.NetFlow{},
				Error(fmt.Sprint("Unable to setup and run client {", c.Label()," ", c.Key(),
					"}. - Unable to send source address. - ", err.Error()))
		}
	}

	// c.connection.Identifier.Src = network.Address{Addr: *backboneConnection.LocalAddr().(*net.TCPAddr)}
	// c.connection.Identifier.Src = network.AddressGenerator{}.New(backboneConnection.LocalAddr().String())
	c.connection.Identifier.Src = backboneConnection.LocalAddr()

	// Create the channels
	c.ingress = make(chan *shila.Packet, c.config.SizeIngressBuff)
	c.egress  = make(chan *shila.Packet, c.config.SizeEgressBuff)

	go c.serveIngress()
	go c.serveEgress()

	log.Verbose.Print("Client {", c.Label(), "} successfully established connection to {", c.Key(), "}.")

	c.isSetup   = true
	c.isRunning = true

	return c.connection.Identifier, nil
}

func (c *Client) TearDown() error {

	log.Verbose.Print("Tear down client {", c.Label(), "}. connecting to {", c.Key(), "}.")

	c.isValid = false
	c.isRunning = false
	c.isSetup = false

	// Close the egress channel
	// Client stops sending out packets
	close(c.egress)

	// Close the connection
	// Stops the ingress processing
	err := c.connection.Backbone.Close()

	// Close the ingress channel
	// Working side no longer processes this endpoint
	close(c.ingress)

	return err
}

func (c *Client) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: c.ingress, Egress: c.egress}
}

func (c *Client) Label() shila.EndpointLabel {
	return c.label
}

func (c *Client) serveIngress() {

	ingressRaw := make(chan byte, c.config.SizeReadBuffer)
	go c.packetize(ingressRaw)

	reader := io.Reader(c.connection.Backbone)
	storage := make([]byte, c.config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, c.config.BatchSizeRead)
		if err != nil && !c.IsValid() {
			// Client is no longer valid, there is no need to try to stay alive.
			close(ingressRaw)
			return
		} else if err != nil {
			// Wait some time, then check if client is still valid (server can close connection earlier..)
			time.Sleep(2 * time.Second) // TODO for SCION: Add to config
			if c.IsValid() {
				panic(fmt.Sprint("Client {", c.Key(), "} unable to read data from backbone connection."))
				// TODO for SCION: Client might still valid, that is, a connection relies on this client! Try to reestablish?
			}
			close(ingressRaw)
			return
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (c *Client) serveEgress() {
	writer := io.Writer(c.connection.Backbone)
	for p := range c.egress {
		_, err := writer.Write(p.Payload)
		if err != nil && !c.IsValid() {
			// Error doesn't matter, client is no longer valid anyway.
			return
		} else if err != nil {
			// Wait some time, then check if client is still valid (server can close connection earlier..)
			time.Sleep(2 * time.Second) // TODO for SCION: Add to config
			if c.IsValid() {
				panic(fmt.Sprint("Client {", c.Key(), "} unable to write data to backbone connection."))
				// TODO for SCION: Client might still valid, that is, a connection relies on this client! Try to reestablish?
			}
		}
	}
}

func (c *Client) packetize(ingressRaw chan byte) {
	for {
		if rawData, _ := tcpip.PacketizeRawData(ingressRaw, c.config.SizeReadBuffer); rawData != nil { // TODO: Handle error
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				panic(fmt.Sprint("Unable to get IP netFlow in packetizer of client {", c.Key(),
					"}. - ", err.Error())) // TODO: Handle panic!
			} else {
				c.ingress <- shila.NewPacket(c, iPHeader, rawData)
			}
		} else {
			return
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
