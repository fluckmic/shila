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

type StateIdentifier uint8
const (
	_                           = iota
	Established StateIdentifier = iota
	ServerReady
	ClientReady
	ClientEstablished
	Closed
	Raw
)

type State struct {
	previous StateIdentifier
	current  StateIdentifier
}

func(s StateIdentifier) String() string {
	switch s {
	case Established: 			return "Established"
	case ServerReady: 			return "ServerReady"
	case ClientReady: 			return "ClientReady"
	case ClientEstablished:		return "ClientEstablished"
	case Closed:				return "Closed"
	case Raw:					return "Raw"
	}
	return "Unknown"
}

func (s *State) Set(newState StateIdentifier) {
	s.previous = s.current
	s.current = newState
}

func (s *State) Current() StateIdentifier {
	return s.current
}

func (s *State) Previous() StateIdentifier {
	return s.previous
}

type Connection struct {
	id          model.IPHeaderKey
	header      model.NetworkHeader
	state       State
	channels    Channels
	lock        sync.Mutex
	touched     time.Time
	sent        bool
	received    bool
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	routing     *model.Mapping
}

type Channels struct {
	KernelEndpoint  model.PacketChannels // Kernel end point
	NetworkEndpoint model.PacketChannels // Network end point
	Contacting      model.PacketChannels // End point for connection establishment
}

func New(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing *model.Mapping,
	id model.IPHeaderKey) *Connection {
	return &Connection{
		id: 			id,
		state: 			State{Raw, Raw},
		lock:			sync.Mutex{},
		touched: 		time.Now(),
		kernelSide:	 	kernelSide,
		networkSide: 	networkSide,
		routing: 		routing,
	}
}

func (c *Connection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Tear down all endpoints possibly associated with this connection
	// TODO: Rethink
	_ = c.networkSide.TeardownContactingClientEndpoint(c.header.Dst)
	_ = c.networkSide.TeardownTrafficSeverEndpoint(c.header.Src)
	_ = c.networkSide.TeardownTrafficClientEndpoint(c.header.Dst, c.header.Path)

	c.state.Set(Closed)
}

func (c *Connection) ProcessPacket(p *model.Packet) error {

	c.lock.Lock()
	defer c.lock.Unlock()

	log.Verbose.Print("Connection {",c.ID(),"} in state {",c.state.Current(),"} " +
		"starts processing packet from {",p.GetIPHeader().Src.IP,"} to {",p.GetIPHeader().Dst.IP,"}.")

	key := p.IPHeaderKey()
	if key != c.id {
		return Error(fmt.Sprint("Cannot process packet - getIPHeader mismatch: ", model.IPHeaderKey(key), " ", c.id, "."))
	}

	// From where was the packet received?
	switch p.GetEntryPoint().Label() {
		case model.KernelEndpoint: 				c.processPacketFromKerep(p)
		case model.ContactingNetworkEndpoint: 	c.processPacketFromContactingEndpoint(p)
		case model.TrafficNetworkEndpoint:		c.processPacketFromTrafficEndpoint(p)
		default: 								return Error(fmt.Sprint("Cannot process packet - Unknown entry device."))
	}

	log.Verbose.Print("Connection {",c.ID(),"} done with processing packet from {",p.GetIPHeader().Src.IP,"} to {",
		p.GetIPHeader().Dst.IP,"}; ending up in state {",c.state.Current(),"}.")

	return nil
}

func (c *Connection) processPacketFromKerep(p *model.Packet) error {
	switch c.state.Current() {
	case Raw:				return c.processPacketFromKerepStateRaw(p)

	case ClientReady:		p.SetNetworkHeader(c.header)
							c.touched = time.Now()
							c.channels.Contacting.Egress <- p
							c.state.Set(ClientReady)
							return nil

	case ServerReady:		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.SetNetworkHeader(c.header)
							c.channels.NetworkEndpoint.Egress <- p
							c.state.Set(ServerReady)
							return nil

	case ClientEstablished:	p.SetNetworkHeader(c.header)
							c.touched = time.Now()
							c.channels.NetworkEndpoint.Egress <- p
							c.state.Set(ClientEstablished)
							return nil

	case Established:		p.SetNetworkHeader(c.header)
							c.touched = time.Now()
							c.channels.NetworkEndpoint.Egress <- p
							c.state.Set(Established)
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							c.state.Set(Closed)
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromContactingEndpoint(p *model.Packet) error {
	switch c.state.Current() {

	case Raw:				return c.processPacketFromContactingEndpointStateRaw(p)

	case ClientReady:		log.Info.Println("Drop packet - Both endpoints in client state.")
							c.state.Set(ClientReady)
							return nil

	case ServerReady:		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state.Set(ServerReady)
							return nil

	case Established: 		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state.Set(Established)
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							c.state.Set(Closed)
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromTrafficEndpoint(p *model.Packet) error {
	switch c.state.Current() {

	case Raw:				c.state.Set(Raw)
							return Error(fmt.Sprint("Cannot process packet - Invalid connection state ", c.state))

							// A packet from the traffic endpoint is just received if network connection is established
	case ClientReady:		c.state.Set(ClientReady)
							return Error(fmt.Sprint("Cannot process packet - Invalid connection state ", c.state))

	case ServerReady:		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state.Set(Established)
							return nil

	case ClientEstablished: // The very first packet received through the traffic endpoint holds the MPTCP receiver key
							// which we need later to be able to get the network destination address for the sub flows.
							if key, ok, err := layer.GetMPTCPSenderKey(p.GetRawPayload()); ok {
								if err == nil {
									if err := c.routing.InsertFromMPTCPEndpointKey(key, c.header.Src, c.header.Dst, c.header.Path); err != nil {
										panic(fmt.Sprint("Error in fetching receiver key in connection {", c.ID(), "}. - ", err.Error())) // TODO: Handle panic!
									}
								} else {
									panic(fmt.Sprint("Error in fetching receiver key in connection {", c.ID(), "} beside " +
										"having packet {", p.GetIPHeader(), "} containing it. - ", err.Error())) // TODO: Handle panic!
								}
							}

							c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state.Set(Established)
							return nil

	case Established: 		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state.Set(Established)
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							c.state.Set(Closed)
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
		c.state.Set(Closed)
		return Error(fmt.Sprint("Error in packet processing. - Invalid entry point type."))
	}

	// Create the packet header which is associated with the connection
	// If the packet contains a receiver token, then the new connection is a MPTCP sub flow.
	if token, ok, err := layer.GetMPTCPReceiverToken(p.GetRawPayload()); ok {
		if err == nil {
			if networkHeader, ok := c.routing.RetrieveFromMPTCPEndpointToken(token); ok {
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
	} else if networkHeader, ok := c.routing.RetrieveFromIPAddressPortKey(p.IPHeaderDstKey()); ok {
		c.header = networkHeader
		p.SetNetworkHeader(networkHeader)
	// No valid option to get network header :(
	} else {
		panic(fmt.Sprint("No valid option to create network header for packet {", p.GetIPHeader(), "}.")) // TODO: Handle panic!
	}

	// Create the contacting connection; srcAddr is the address source address of the contacting backbone connection
	var srcAddrContacting model.NetworkAddress
	if channels, srcAddr, err := c.networkSide.EstablishNewContactingClientEndpoint(c.header.Dst); err != nil {
		c.state.Set(Closed)
		return Error(fmt.Sprint("Error in packet processing. - ", err.Error()))
	} else {
		c.channels.Contacting = channels
		srcAddrContacting = srcAddr
	}

	// Send the packet via the contacting channel
	c.touched = time.Now()
	c.channels.Contacting.Egress <- p

	// Initiate try connect to address via path
	go func() {
		// TODO: Upon the connection, the traffic client endpoint informs the traffic server on behalf of whom it initiated the connection.
		networkHeaderTraffic := model.NetworkHeader{Src: srcAddrContacting, Path: c.header.Path, Dst: c.header.Dst}
		if channels, err := c.networkSide.EstablishNewTrafficClientEndpoint(networkHeaderTraffic); err != nil {
			c.Close()
			log.Info.Print("Unable to establish traffic connection. - ", err.Error())
		} else {
			c.channels.NetworkEndpoint = channels
			c.lock.Lock()
			defer c.lock.Unlock()
			c.state.Set(ClientEstablished)
			// The contacting client endpoint is no longer needed.
			_ = c.networkSide.TeardownContactingClientEndpoint(c.header.Dst)
			log.Verbose.Print("Established traffic connection to {",c.header.Dst," via ",c.header.Path,"}.")
		}
	}()

	// Set new state
	c.state.Set(ClientReady)

	return nil
}

func (c *Connection) processPacketFromContactingEndpointStateRaw(p *model.Packet) error {

	// Get the kernel endpoint from the kernel side manager
	dstKey := p.IPHeaderDstIPKey()
	if channels, ok := c.kernelSide.GetTrafficChannels(dstKey); ok {
		c.channels.KernelEndpoint = channels
	} else {
		c.state.Set(Closed)
		return Error(fmt.Sprint("Cannot process packet - No kernel endpoint for ", dstKey))
	}

	// Send packet to kernel endpoint
	// --> 	Could still be that connection cannot be established, since we have no idea if there is actually a server listening
	c.channels.KernelEndpoint.Egress <- p

	// If the packet is received through the contacting endpoint (server), then it's network header
	// is already set. This is the responsibility of the corresponding network server implementation.
	c.header = p.GetNetworkHeader()

	// Request new incoming connection from network side.
	// ! The receiving network endpoint is responsible to correctly set the destination network address! !
	if channels, err := c.networkSide.EstablishNewServerEndpoint(c.header); err != nil {
		c.state.Set(Closed)
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	} else {
		c.channels.NetworkEndpoint = channels
	}

	// Set new state
	c.state.Set(ServerReady)

	return nil
}

func (c *Connection) ID() string {
	return string(c.id)
}