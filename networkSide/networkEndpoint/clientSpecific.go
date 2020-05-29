package networkEndpoint

import "C"
import (
	"fmt"
	"io"
	"net"
	"shila/core/shila"
	"shila/layer/tcpip"
	"shila/log"
	"shila/networkSide/network"
	"time"
)

var _ shila.NetworkClientEndpoint = (*Client)(nil)

type Client struct {
	Base
	connection networkConnection
}

type networkConnection struct {
	Identifier shila.NetFlow
	Backbone   *net.TCPConn
}

func NewClient(netConnId shila.NetFlow, label shila.EndpointLabel) shila.NetworkClientEndpoint {
	return &Client{
		Base: 				Base{
								label: label,
								state: shila.NewEntityState(),
							},
		connection:		    networkConnection{Identifier: netConnId},
	}
}

func (c *Client) SetupAndRun() (shila.NetFlow, error) {

	if c.state.Not(shila.Uninitialized) {
		return shila.NetFlow{}, shila.CriticalError(fmt.Sprint("Entity in wrong state {", c.state, "}."))
	}

	// Establish a connection to the server endpoint
	dst := c.connection.Identifier.Dst.(*net.TCPAddr)
	backboneConnection, err := net.DialTCP(dst.Network(), nil, dst)
	if err != nil {
		if c.Label() == shila.TrafficNetworkEndpoint {
			err = shila.TolerableError(err.Error())
		} else {
			// For a contacting endpoint, the issue is most likely that there is no one listening..
			err = shila.ThirdPartyError(err.Error())
		}
		return shila.NetFlow{}, err
	}
	c.connection.Backbone = backboneConnection

	if c.Label() == shila.TrafficNetworkEndpoint {

		// Before setting the own src address, a traffic client sends the currently set src address to the server;
		// which should be (or is.) the src address of the corresponding contacting client endpoint. This information
		// is required to be able to do the mapping on the server side.
		if _, err := c.connection.Backbone.Write([]byte(fmt.Sprintln(c.connection.Identifier.Src.String()))); err != nil {
			return shila.NetFlow{}, shila.TolerableError(err.Error())
		}
	}

	c.connection.Identifier.Src = backboneConnection.LocalAddr()

	// Create the channels
	c.ingress = make(chan *shila.Packet, Config.SizeIngressBuffer)
	c.egress  = make(chan *shila.Packet, Config.SizeEgressBuffer)

	go c.serveIngress()
	go c.serveEgress()

	c.state.Set(shila.Running)
	log.Verbose.Print("Client {", c.Label(), "} successfully established connection to {", c.Key(), "}.")
	return c.connection.Identifier, nil
}

func (c *Client) Key() shila.EndpointKey {
	path, _ := network.PathGenerator{}.New("")
	return shila.EndpointKey(shila.GetNetworkAddressAndPathKey(c.connection.Identifier.Dst, path))
}

func (c *Client) TearDown() error {

	log.Verbose.Print("Tear down client {", c.Label(), "} connecting to {", c.Key(), "}.")

	c.state.Set(shila.TornDown)

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

	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go c.packetize(ingressRaw)

	reader := io.Reader(c.connection.Backbone)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		if err != nil {
			// Wait a little bit, the server was might earlier with
			// regularly closing the connection.
			time.Sleep(Config.WaitingTimeUntilConnectionRetry)
			if c.state.Not(shila.Running) {
				close(ingressRaw)
				return
			}
			// TODO: https://github.com/fluckmic/shila/issues/14
			log.Info.Print("Client {", c.Key(), "} unable to write data to backbone connection.")
			panic("No reconnection functionality implemented.")
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
		if err != nil {
			// Wait a little bit, the server was might earlier with
			// regularly closing the connection.
			time.Sleep(Config.WaitingTimeUntilConnectionRetry)
			if c.state.Not(shila.Running) {
				return
			}
			// TODO: https://github.com/fluckmic/shila/issues/14
			log.Info.Print("Client {", c.Key(), "} unable to write data to backbone connection.")
			panic("No reconnection functionality implemented.")
		}
	}
}

func (c *Client) packetize(ingressRaw chan byte) {
	for {
		if rawData, _ := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil { // TODO: Handle error
			if ipFlow, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP flow from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print("Unable to get IP net flow in packetizer of client {", c.Key(),	"}. - ", err.Error())
			} else {
				c.ingress <- shila.NewPacket(c, ipFlow, rawData)
			}
		} else {
			return
		}
	}
}