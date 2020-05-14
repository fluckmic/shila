package connection

import (
	"fmt"
	"shila/core/model"
	"shila/kernelSide"
	"shila/kernelSide/kernelEndpoint"
	"shila/layer"
	"shila/log"
	"shila/networkSide"
	"sync"
	"time"
)

type State uint8
const (
	_                 = iota
	Established State = iota
	ServerReady
	ClientReady
	Closed
	Raw
)

type Type uint8
const (
	_				= iota
	Unknown Type	= iota
	MainFlow
	SubFlow
)

type Connection struct {
	id          model.IPHeaderKey
	header      model.NetworkHeader
	state       State
	channels    Channels
	lock        sync.Mutex
	touched     time.Time
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	routing 	*model.Mapping
}

type Channels struct {
	KernelEndpoint  model.TrafficChannels // Kernel end point
	NetworkEndpoint model.TrafficChannels // Network end point
	Contacting      model.TrafficChannels // End point for connection establishment
}

func New(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing *model.Mapping,
	id model.IPHeaderKey) *Connection {
	return &Connection{id: id, header: model.NetworkHeader{} , state: Raw,
		touched: time.Now(), kernelSide: kernelSide, networkSide: networkSide, routing: routing}
}

func (c *Connection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Tear down all endpoints possibly associated with this connection
	// TODO: Clean up depends on the "side" of the connection.
	/*
	_ = c.networkSide.TeardownContactingClientEndpoint(c.header.Dst)
	_ = c.networkSide.TeardownTrafficSeverEndpoint(c.header.)
	_ = c.networkSide.TeardownTrafficClientEndpoint(c.header.Dst, c.header.Path)
	*/

	c.state = Closed
}

func (c *Connection) ProcessPacket(p *model.Packet) error {

	c.lock.Lock()
	defer c.lock.Unlock()

	key := p.IPHeaderKey()
	if key != c.id {
		return Error(fmt.Sprint("Cannot process packet - getIPHeader mismatch: ", model.IPHeaderKey(key), " ", c.id, "."))
	}

	// From where was the packet received?
	switch p.GetEntryPoint().Label() {
		case model.KernelEndpoint: 				return c.processPacketFromKerep(p)
		case model.ContactingNetworkEndpoint: 	return c.processPacketFromContactingEndpoint(p)
		case model.TrafficNetworkEndpoint:		return c.processPacketFromTrafficEndpoint(p)
		default: 								return Error(fmt.Sprint("Cannot process packet - Unknown entry device."))
	}

	return nil
}

func (c *Connection) processPacketFromKerep(p *model.Packet) error {
	switch c.state {
	case Raw:				return c.processPacketFromKerepStateRaw(p)

	case ClientReady:		p.SetNetworkHeader(c.header)
							c.touched = time.Now()
							c.channels.Contacting.Egress <- p
							return nil

	case ServerReady:		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.SetNetworkHeader(c.header)
							c.channels.NetworkEndpoint.Egress <- p
							return nil

	case Established: 		p.SetNetworkHeader(c.header)
							c.touched = time.Now()
							c.channels.NetworkEndpoint.Egress <- p
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromContactingEndpoint(p *model.Packet) error {
	switch c.state {

	case Raw:				return c.processPacketFromContactingEndpointStateRaw(p)

	case ClientReady:		log.Info.Println("Drop packet - Both endpoints in client state.")
							return nil

	case ServerReady:		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							return nil

	case Established: 		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromTrafficEndpoint(p *model.Packet) error {
	switch c.state {

	case Raw:				return Error(fmt.Sprint("Cannot process packet - Invalid connection state ", c.state))

							// A packet from the traffic endpoint is just received if network connection is established
	case ClientReady:		return Error(fmt.Sprint("Cannot process packet - Invalid connection state ", c.state))

	case ServerReady:		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state = Established
							return nil

	case Established: 		c.touched = time.Now()
							// In the very first message received on the client side in a MPTCP main flow
							// the packet contains the receivers key (IPHeaderKey-B). This key is later needed to
							// be able to find the network destination address and path for a subflow
							// TODO.
							c.channels.KernelEndpoint.Egress <- p
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromKerepStateRaw(p *model.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.GetEntryPoint()
	if entryPoint, ok := ep.(*kernelEndpoint.Device); ok {
		c.channels.KernelEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from kernel end point
		c.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards kernel end point
	} else {
		c.state = Closed
		return Error(fmt.Sprint("Error in packet processing. - Invalid entry point type."))
	}

	// Create the packet header which is associated with the connection
	// If the packet contains a receiver token, then the new connection is a MPTCP sub flow.
	if token, ok, err := layer.GetMPTCPReceiverToken(p.GetRawPayload()); ok {
		if err == nil {
			if networkHeader, ok := c.routing.RetrieveFromReceiverToken(token); ok {
				c.header = networkHeader
				p.SetNetworkHeader(networkHeader)
			} else {
				panic(fmt.Sprint("No network header available in routing for receiver token {", token,"} beside " +
					"having packet {", p.GetIPHeader(), "} containing it.")) // TODO: Handle panic!
			}
		} else {
			panic(fmt.Sprint("Unable to get receiver token for packet {", p.GetIPHeader(), "}. - ", err.Error())) // TODO: Handle panic!
		}
	// For a MPTCP main flow the network header can probably be extracted from the IP options
	} else if networkHeader, ok, err := layer.GetNetworkHeaderFromIPOptions(p.GetRawPayload()); ok {
		if err == nil {
			c.header = networkHeader
			p.SetNetworkHeader(networkHeader)
		} else {
			panic(fmt.Sprint("Unable to get IP options for packet {", p.GetIPHeader(), "}. - ", err.Error())) // TODO: Handle panic!
		}
	// For a MPTCP main flow the network header is probably available in the routing table
	} else if networkHeader, ok := c.routing.RetrieveFromIPAddressKey(p.IPHeaderDstKey()); ok {
		c.header = networkHeader
		p.SetNetworkHeader(networkHeader)
	// No valid option to get network header :(
	} else {
		panic(fmt.Sprint("No valid option to create network header for packet {", p.GetIPHeader(), "}.")) // TODO: Handle panic!
	}

	// Create the contacting connection
	if channels, err := c.networkSide.EstablishNewContactingClientEndpoint(c.header.Dst); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Error in packet processing. - ", err.Error()))
	} else {
		c.channels.Contacting = channels
	}

	// Send the packet via the contacting channel
	c.touched = time.Now()
	c.channels.Contacting.Egress <- p

	// Initiate try connect to address via path
	go func() {
		if channels, err := c.networkSide.EstablishNewTrafficClientEndpoint(c.header.Dst, c.header.Path); err != nil {
			c.Close()
			log.Info.Print("Unable to establish traffic connection. - ", err.Error())
		} else {
			c.channels.NetworkEndpoint = channels
			c.lock.Lock()
			defer c.lock.Unlock()
			c.state = Established
			// The contacting client endpoint is no longer needed.
			_ = c.networkSide.TeardownContactingClientEndpoint(c.header.Dst)
			log.Verbose.Print("Established traffic connection from ", c.header.Src, " to ", c.header.Dst, " via ", c.header.Path, ".")
		}
	}()

	// Set new state
	c.state = ClientReady

	return nil
}

func (c *Connection) processPacketFromContactingEndpointStateRaw(p *model.Packet) error {

	// Get the kernel endpoint from the kernel side manager
	dstKey := p.IPHeaderDstKey()
	if channels, ok := c.kernelSide.GetTrafficChannels(dstKey); ok {
		c.channels.KernelEndpoint = channels
	} else {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - No kernel endpoint for ", dstKey))
	}

	// Send packet to kernel endpoint
	// --> 	Could still be that connection cannot be established, since we have no idea if there is actually a server listening
	c.channels.KernelEndpoint.Egress <- p

	/*
	// Create the packet header which is associated with the connection
	if header, err := p.GetNetworkHeader(c.routing); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	} else {
		c.header = header
	}
	*/

	// Request new incoming connection from network side.
	// ! The receiving network endpoint is responsible to correctly set the destination network address! !
	if channels, err := c.networkSide.EstablishNewServerEndpoint(c.header.Dst); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	} else {
		c.channels.NetworkEndpoint = channels
	}

	// Set new state
	c.state = ServerReady

	return nil
}

func (c *Connection) ID() string {
	return string(c.id)
}