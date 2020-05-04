package connection

import (
	"fmt"
	"shila/kersi"
	"shila/kersi/kerep"
	"shila/log"
	"shila/networkSide"
	"shila/networkSide/networkEndpoint"
	"shila/shila"
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

type Connection struct {
	id		 	ID
	header   	shila.PacketHeader
	state 	 	State
	channels 	Channels
	lock  	 	sync.Mutex
	touched  	time.Time
	kernelSide 	kersi.Manager
	networkSide networkSide.Manager
}

type Channels struct {
	KernelEndpoint  shila.TrafficChannels 	// Kernel end point
	NetworkEndpoint shila.TrafficChannels 	// Network end point
	Contacting      shila.ContactingChannel	// End point for connection establishment
}

func New(kernelSide kersi.Manager, networkSide networkSide.Manager, id ID) *Connection {
	return &Connection{id, shila.PacketHeader{} ,Raw, Channels{} ,
		sync.Mutex{}, time.Now(), kernelSide, networkSide}
}

func (c *Connection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.state = Closed
}

func (c *Connection) ProcessPacket(p *shila.Packet) error {

	c.lock.Lock()
	defer c.lock.Unlock()

	key, err := p.Key()
	if err != nil {
		return err
	}
	if ID(key) != c.id {
		return Error(fmt.Sprint("Cannot process packet - ID mismatch: ", ID(key), " ", c.id, "."))
	}

	// From where was the packet received?
	switch p.EntryPoint().Label() {
		case shila.KernelEndpoint: 				return c.processPacketFromKerep(p)
		case shila.ContactingNetworkEndpoint: 	return c.processPacketFromContactingEndpoint(p)
		case shila.TrafficNetworkEndpoint:		return c.processPacketFromTrafficEndpoint(p)
		default: 								return Error(fmt.Sprint("Cannot process packet - Unknown entry device."))
	}

	return nil
}

func (c *Connection) processPacketFromKerep(p *shila.Packet) error {

	switch c.state {

	case Raw:				return c.processPacketFromKerepStateRaw(p)

	case ClientReady:		c.assignPacketHeaderFromConnectionToPacket(p)
							c.touched = time.Now()
							c.channels.Contacting.Channels.Egress <- p
							return nil

	case ServerReady:		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							c.assignPacketHeaderFromConnectionToPacket(p)
							c.channels.NetworkEndpoint.Egress <- p
							return nil

	case Established: 		c.assignPacketHeaderFromConnectionToPacket(p)
							c.touched = time.Now()
							c.channels.NetworkEndpoint.Egress <- p
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromContactingEndpoint(p *shila.Packet) error {

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

func (c *Connection) processPacketFromTrafficEndpoint(p *shila.Packet) error {
	return nil
}

func (c *Connection) processPacketFromKerepStateRaw(p *shila.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.EntryPoint()
	if entryPoint, ok := ep.(*kerep.Device); ok {
		c.channels.KernelEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from kernel end point
		c.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards kernel end point
	} else {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - Invalid entry point type."))
	}

	// Create the packet header which is associated with the connection
	if err := c.createPacketHeader(p); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	}

	// Create the contacting connection, which is part of the connection itself.
	if err := c.createAndEstablishContactingChannel(); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	}

	// Send the packet via the contacting channel
	c.assignPacketHeaderFromConnectionToPacket(p)
	c.touched = time.Now()
	c.channels.Contacting.Channels.Egress <- p

	// Initiate try connect to address via path
	if err := c.createAndInitiateEstablishmentOfTrafficChannel(); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	}

	// Set new state
	c.state = ClientReady

	return nil
}

func (c *Connection) ID() string {
	return string(c.id)
}

func (c *Connection) createAndEstablishContactingChannel() error {

	// Request from network side!?

	// TODO: Make size of channels configurable
	c.channels.Contacting.Channels.Egress =  make(shila.PacketChannel, 10)
	c.channels.Contacting.Channels.Ingress = make(shila.PacketChannel, 10)
	// TODO: Maybe assign a "default" path here, since this is the connection used for the contacting
	c.channels.Contacting.Endpoint = networkEndpoint.Generator{}.NewClient(c.header.Dst, c.header.Path, shila.ContactingNetworkEndpoint,
		shila.TrafficChannels{Ingress: c.channels.Contacting.Channels.Egress, Egress: c.channels.Contacting.Channels.Ingress})

	if err := c.channels.Contacting.Endpoint.SetupAndRun(); err != nil {
		return Error(fmt.Sprint("Failed to establish contacting channel - ", err.Error()))
	}

	return nil
}

func (c *Connection) assignPacketHeaderFromConnectionToPacket(p *shila.Packet) {
	p.PacketHeader().Src  = c.header.Src
	p.PacketHeader().Path = c.header.Path
	p.PacketHeader().Dst  = c.header.Dst
}

func (c *Connection) createPacketHeader(p *shila.Packet) error {
	// TODO:
	return nil
}

func (c *Connection) createAndInitiateEstablishmentOfTrafficChannel() error {
	// TODO:

	// Stop and clean up connection try when connection state is Closed
	// Reset timer and set state to Established when connection try succeeds
	return nil
}

func (c *Connection) processPacketFromContactingEndpointStateRaw(p *shila.Packet) error {

	/*
	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.EntryPoint()
	if entryPoint, ok := ep.(*networkEndpoint.Server); ok {
		c.channels.NetworkEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from network end point
		c.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards network end point
	} else {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - Invalid entry point type."))
	} */

	// Get the kernel endpoint from the kernel side manager
		// --> error if there is no such kernel endpoint (dest ip)
		// --> assign the kernel endpoint channels

	// Send packet to kernel endpoint
		// --> 	Could still be that connection cannot be established, since we have no idea if there
		//		is actually a server listening

	// Request new incoming connection from network side.

	// Set new state
	c.state = ServerReady

	return nil
}