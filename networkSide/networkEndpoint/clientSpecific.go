//
package networkEndpoint

import (
	"encoding/gob"
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
	Identifier shila.Flow
	Backbone   *net.TCPConn
}

func NewClient(flow shila.Flow, label shila.EndpointLabel, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return &Client{
		Base: 				Base{
								label: 			label,
								state: 			shila.NewEntityState(),
								endpointIssues: endpointIssues,
							},
		connection:		    networkConnection{Identifier: flow},
	}
}

func (c *Client) SetupAndRun() (shila.NetFlow, error) {

	if c.state.Not(shila.Uninitialized) {
		return shila.NetFlow{}, shila.CriticalError(fmt.Sprint("Entity in wrong state {", c.state, "}."))
	}

	// Backup the src network address of the corresponding contacting endpoint (in case of a traffic network endpoint)
	srcAddrContacting := c.connection.Identifier.NetFlow.Src.(*net.TCPAddr)

	// Establish a connection to the server endpoint
	dst := c.connection.Identifier.NetFlow.Dst.(*net.TCPAddr)
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
	c.connection.Identifier.NetFlow.Src = backboneConnection.LocalAddr()

	log.Verbose.Print(c.message("Established connection."))

	// Send the IP flow to the server
	writer := io.Writer(c.connection.Backbone)
	encoder := gob.NewEncoder(writer)
	if err := encoder.Encode(c.connection.Identifier.IPFlow); err != nil {
		return shila.NetFlow{}, shila.PrependError(err, "Failed to transmit IP flow.")
	}

	if c.Label() == shila.TrafficNetworkEndpoint {
		// Before setting the own src address, a traffic client sends the currently set src address to the server;
		// which should be (or is.) the src address of the corresponding contacting client endpoint. This information
		// is required to be able to do the mapping on the server side.
		encoder := gob.NewEncoder(writer)
		if err := encoder.Encode(srcAddrContacting); err != nil {
			return shila.NetFlow{}, shila.PrependError(err, "Failed to transmit src network address.")
		}
		log.Verbose.Print(c.message(fmt.Sprint("Sent source address {", srcAddrContacting, "} of corresponding contact client endpoint.")))
	}
	
	// Create the channels
	c.ingress = make(chan *shila.Packet, Config.SizeIngressBuffer)
	c.egress  = make(chan *shila.Packet, Config.SizeEgressBuffer)

	go c.serveIngress()
	go c.serveEgress()

	c.state.Set(shila.Running)

	return c.connection.Identifier.NetFlow, nil
}

func (c *Client) Key() shila.EndpointKey {
	path, _ := network.PathGenerator{}.New("")
	return shila.EndpointKey(shila.GetNetworkAddressAndPathKey(c.connection.Identifier.NetFlow.Dst, path))
}

func (c *Client) TearDown() error {

	log.Verbose.Print(c.message("Got torn down."))

	c.state.Set(shila.TornDown)

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

			// For the moment; we just tear down the whole client if there is an issue with the backbone connection.
			c.endpointIssues <- shila.EndpointIssuePub{
				Publisher: 	c,
				Flow:		c.connection.Identifier,
				Error:    	shila.ThirdPartyError("Unable to read data."),
			}
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
		if err != nil {
			// Wait a little bit, the server was might earlier with
			// regularly closing the connection.
			time.Sleep(Config.WaitingTimeUntilConnectionRetry)
			if c.state.Not(shila.Running) {
				return
			}
			// TODO: https://github.com/fluckmic/shila/issues/14

			// For the moment; we just tear down the whole client if there is an issue with the backbone connection.
			c.endpointIssues <- shila.EndpointIssuePub{
				Publisher: 	c,
				Flow:		c.connection.Identifier,
				Error:     	shila.ThirdPartyError("Unable to write data."),
			}
			return
		}
	}
}

func (c *Client) packetize(ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
			if ipFlow, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP flow from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print("Unable to get IP net flow in packetizer of client {", c.Key(),	"}. - ", err.Error())
			} else {
				c.ingress <- shila.NewPacket(c, ipFlow, rawData)
			}
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			c.endpointIssues <- shila.EndpointIssuePub{
				Publisher: 	c,
				Flow: 		c.connection.Identifier,
				Error:     	shila.PrependError(err, "Error in raw data packetizer."),
			}
			return
		}
	}
}

func (c *Client) Flow() shila.Flow {
	return c.connection.Identifier
}

func (c *Client) message(s string) string {
	return fmt.Sprint("Client {",c.Label(), " - ", c.connection.Identifier.NetFlow.Src.String()," -> ",
		c.connection.Identifier.NetFlow.Dst.String(),"}: ", s)
}