//
package networkEndpoint

import (
	"encoding/gob"
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"io"
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/log"
	"shila/measurements"
	"time"
)

var _ shila.NetworkClientEndpoint = (*Client)(nil)

type Client struct {
	Base
	key             shila.TCPFlowKey
	rConn           *snet.Conn
	tcpFlow         shila.TCPFlow
	netFlow         shila.NetFlow
	lAddrContactEnd shila.NetworkAddress 	// Just set for traffic client network endpoint
}

func NewContactClient(rAddr shila.NetworkAddress, path shila.NetworkPath, tcpFlow shila.TCPFlow, issues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return &Client{
		Base: 						Base{
										Role:    shila.ContactNetworkEndpoint,
										Ingress: make(chan *shila.Packet, config.Config.NetworkEndpoint.SizeIngressBuffer),
										Egress:  make(chan *shila.Packet, config.Config.NetworkEndpoint.SizeIngressBuffer),
										State:   shila.NewEntityState(),
										Issues:  issues,
									},
		tcpFlow: tcpFlow,
		netFlow: shila.NetFlow{Dst: rAddr, Path: path},
	}
}

func NewTrafficClient(lAddrContactEnd shila.NetworkAddress, rAddr shila.NetworkAddress, path shila.NetworkPath, tcpFlow shila.TCPFlow,
	issues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {

	client := NewContactClient(rAddr, path, tcpFlow, issues)

	client.(*Client).Base.Role 		 = shila.TrafficNetworkEndpoint
	client.(*Client).lAddrContactEnd = lAddrContactEnd

	return client
}

func (client *Client) SetupAndRun() (netFlow shila.NetFlow, err error) {

	if client.State.Not(shila.Uninitialized) {
		err = shila.CriticalError(fmt.Sprint("Entity in wrong State ", client.State, "."))
		return
	}

	// Establish a connection.
	if err = client.establishConnection(); err != nil {
		return
	}

	// Send the control message.
	if err = client.sendControlMessage(); err != nil {
		return
	}

	// Start the ingress and egress machinery.
	go client.serveIngress()
	go client.serveEgress()

	client.State.Set(shila.Running)

	return client.netFlow, nil
}

func (client *Client) establishConnection() (err error) {

	scionAddr := client.netFlow.Dst.(*snet.UDPAddr) // FIXME: cast!

	if client.netFlow.Path != nil {
		scionPath := client.netFlow.Path.(snet.Path)	// FIXME: cast!
		appnet.SetPath(scionAddr, scionPath)
	}

	client.rConn, err = appnet.DialAddr(scionAddr)
	if err != nil {
		err = shila.PrependError(ConnectionError(err.Error()), "Cannot establish connection.")
		return
	}

	client.netFlow.Src = client.rConn.LocalAddr().(*net.UDPAddr)	// FIXME: cast!
	log.Verbose.Print(client.Says("Established connection."))

	return
}

func (client *Client) TearDown() error {

	client.State.Set(shila.TornDown)

	err := client.rConn.Close() 		// Close the connection (stops the Ingress processing)
	close(client.Ingress)               // Close the Ingress channel (Working side no longer processes this endpoint)

	log.Verbose.Print(client.Says("Got torn down."))
	return err
}

func (client *Client) Role() shila.EndpointRole {
	return client.Base.Role
}

func (client *Client) 	Identifier() string {
	return fmt.Sprint("Client ", client.Role(), " (", client.netFlow.Src, " -> ", client.netFlow.Dst, ")")
}

func (client *Client) Says(str string) string {
	return  fmt.Sprint(client.Identifier(), ": ", str)
}

func (client *Client) Key() shila.TCPFlowKey {
	return client.key
}

func (client *Client) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: client.Ingress, Egress: client.Egress}
}

func (client *Client) serveIngress() {
	for {
		var pyldMsg payloadMessage
		if err := gob.NewDecoder(client.rConn).Decode(&pyldMsg); err != nil {
			go client.handleConnectionIssue(err)
			// After an issue, we no longer serve ingress. Connection will shut down the client later.
			return
		}
		if len(pyldMsg.Payload) == 0 {
			// From time to to we get a zero payload packet...?
			//log.Error.Println(client.Says("Received zero payload packet."))
		}
		client.Ingress <-  shila.NewPacket(client, client.tcpFlow, pyldMsg.Payload)

	}
}

func (client *Client) serveEgress() {
	for p := range client.Egress {
		err := client.sendPayloadMessage(p.Payload)
		if err != nil {
			go client.handleConnectionIssue(err)
			// After an issue, we no longer server egress. Connection will shut down the client later.
			return
		}
	}
}

func (client *Client) sendPayloadMessage(payload []byte) error {

	// Craft the payload message,..
	pyldMsg := payloadMessage{
		Payload: payload,
	}

	go func() {
		// ...probably create a timestamp for it..
		if config.Config.Logging.DoEgressTimestamping {
			measurements.LogEgressTimestamp(payload)
		}
	}()

	//  ..and encode and send it.
	if err := gob.NewEncoder(io.Writer(client.rConn)).Encode(pyldMsg); err != nil {
		return shila.PrependError(err, "Cannot encode payload message.")
	}

	return nil
}



func (client *Client) sendControlMessage() error {

	// Craft the control message, encode and send it.
	var ctrlMsg controlMessage
	if client.Role() == shila.ContactNetworkEndpoint {
		ctrlMsg = controlMessage{TcpFlow: client.tcpFlow}
	}
	if client.Role() == shila.TrafficNetworkEndpoint {
		ctrlMsg = controlMessage{TcpFlow: client.tcpFlow, LAddrContactEnd: *client.lAddrContactEnd.(*net.UDPAddr)}
	}

	if err := gob.NewEncoder(io.Writer(client.rConn)).Encode(ctrlMsg); err != nil {
		return shila.PrependError(err, "Cannot encode control message.")
	}
	
	return nil
}

func (client *Client) handleConnectionIssue(err error) {

	// Wait a little bit - maybe the client is going to die anyway.
	time.Sleep(time.Second * time.Duration(config.Config.NetworkEndpoint.WaitingTimeAfterConnectionIssue))
	if client.State.Is(shila.Running) {

		if opErr, ok := err.(*snet.OpError); ok {
			client.Issues <- shila.EndpointIssuePub{ Issuer: client, Key: client.Key(), Error: opErr }
		}

		client.Issues <- shila.EndpointIssuePub{ Issuer: client, Key: client.Key(), Error: ConnectionError(err.Error()) }
	}
}
