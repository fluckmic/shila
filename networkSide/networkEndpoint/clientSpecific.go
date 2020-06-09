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
	NetworkEndpointBase
	connection networkConnection
}

func NewClient(flow shila.Flow, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return &Client{
		NetworkEndpointBase: NetworkEndpointBase{
								role:    role,
								ingress: make(chan *shila.Packet, Config.SizeIngressBuffer),
								egress:  make(chan *shila.Packet, Config.SizeIngressBuffer),
								state:   shila.NewEntityState(),
								issues:  issues,
							},
		connection:		    networkConnection{RepresentingFlow: flow},
	}
}

func (c *Client) SetupAndRun() (shila.NetFlow, error) {

	if c.state.Not(shila.Uninitialized) {
		return shila.NetFlow{}, shila.CriticalError(fmt.Sprint("Entity in wrong state {", c.state, "}."))
	}

	// Backup the src network address of the corresponding contacting endpoint (in case of a traffic network endpoint)
	srcAddrContacting := *c.connection.RepresentingFlow.NetFlow.Src.(*net.TCPAddr)
	// FIXME: Also contains the path taken?

	// Establish a connection to the server endpoint
	dst := c.connection.RepresentingFlow.NetFlow.Dst.(*net.TCPAddr)
	backboneConnection, err := net.DialTCP(dst.Network(), nil, dst)
	if err != nil {
		if c.Role() == shila.TrafficNetworkEndpoint {
			err = shila.TolerableError(err.Error())
		} else {
			// For a contacting endpoint, the issue is most likely that there is no one listening..
			err = shila.ThirdPartyError(err.Error())
		}
		return shila.NetFlow{}, err
	}
	c.connection.Backbone = backboneConnection
	c.connection.RepresentingFlow.NetFlow.Src = backboneConnection.LocalAddr()

	log.Verbose.Print(c.message("Established connection."))

	// Send the control msg to the server
	ctrlMsg := controlMessage{
		IPFlow:   				c.connection.RepresentingFlow.IPFlow,
		SrcAddrContactEndpoint: srcAddrContacting,
	}
	if err := gob.NewEncoder(io.Writer(c.connection.Backbone)).Encode(ctrlMsg); err != nil {
		return shila.NetFlow{}, shila.PrependError(err, "Failed to transmit control message.")
	}

	go c.serveIngress()
	go c.serveEgress()

	c.state.Set(shila.Running)

	return c.connection.RepresentingFlow.NetFlow, nil
}

func (c *Client) Key() shila.EndpointKey {
	path, _ := network.PathGenerator{}.New("")
	return shila.EndpointKey(shila.GetNetworkAddressAndPathKey(c.connection.RepresentingFlow.NetFlow.Dst, path))
}

func (c *Client) TearDown() error {

	c.state.Set(shila.TornDown)

	err := c.connection.Backbone.Close()	// Close the connection (stops the ingress processing)
	close(c.ingress) 						// Close the ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(c.message("Got torn down."))
	return err
}

func (c *Client) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: c.ingress, Egress: c.egress}
}

func (c *Client) Role() shila.EndpointRole {
	return c.role
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
			c.issues <- shila.EndpointIssuePub{
				Issuer: c,
				Flow:   c.connection.RepresentingFlow,
				Error:  shila.ThirdPartyError("Unable to read data."),
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
			c.issues <- shila.EndpointIssuePub{
				Issuer: c,
				Flow:   c.connection.RepresentingFlow,
				Error:  shila.ThirdPartyError("Unable to write data."),
			}
			return
		}
	}
}

func (c *Client) packetize(ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
			if ipFlow, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP lAddress from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print("Unable to get IP net lAddress in packetizer of client {", c.Key(),	"}. - ", err.Error())
			} else {
				c.ingress <- shila.NewPacket(c, ipFlow, rawData)
			}
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			c.issues <- shila.EndpointIssuePub{
				Issuer: c,
				Flow:   c.connection.RepresentingFlow,
				Error:  shila.PrependError(err, "Error in raw data packetizer."),
			}
			return
		}
	}
}

func (c *Client) Flow() shila.Flow {
	return c.connection.RepresentingFlow
}

func (c *Client) message(s string) string {
	return fmt.Sprint("Client {",c.Role(), " - ", c.connection.RepresentingFlow.NetFlow.Src.String()," -> ",
		c.connection.RepresentingFlow.NetFlow.Dst.String(),"}: ", s)
}