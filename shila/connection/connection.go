package connection

import (
	"fmt"
	"shila/kersi/kerep"
	"shila/log"
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
	id		 ID
	header   shila.PacketHeader
	state 	 State
	channels Channels
	lock  	 sync.Mutex
	touched  time.Time
}

type Channels struct {
	KernelEndpoint  shila.TrafficChannels 	// Kernel end point
	NetworkEndpoint shila.TrafficChannels 	// Network end point
	Contacting      shila.ContactingChannel	// End point for connection establishment
}

func New(id ID) *Connection {
	return &Connection{id, shila.PacketHeader{} ,Raw, Channels{} ,sync.Mutex{}, time.Now()}
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
		case shila.KernelEndpoint: 	return c.processPacketFromKerep(p)
		default: 					return Error(fmt.Sprint("Cannot process packet - Unknown entry device."))
	}

	return nil
}

func (c *Connection) processPacketFromKerep(p *shila.Packet) error {

	switch c.state {
	case Raw:			return c.processPacketFromKerepStateRaw(p)
	case ClientReady:	return c.processPacketFromKerepStateClientReady(p)
	case ServerReady:	return c.processPacketFromKerepStateServerReady(p)
	case Established: 	return c.processPacketFromKerepStateEstablished(p)
	case Closed: 		return c.processPacketFromKerepStateClosed(p)
	default: 			return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}

	return nil
}

func (c *Connection) processPacketFromKerepStateRaw(p *shila.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.EntryPoint()
	if entryPoint, ok := ep.(*kerep.Device); ok {
		c.channels.KernelEndpoint.Ingress = entryPoint.Channels.Ingress // ingress from kernel end point
		c.channels.KernelEndpoint.Egress  = entryPoint.Channels.Egress  // egress towards kernel end point
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

func (c *Connection) processPacketFromKerepStateClientReady(p *shila.Packet) error {
	c.assignPacketHeaderFromConnectionToPacket(p)
	c.touched = time.Now()
	c.channels.Contacting.Channels.Egress <- p
	return nil
}

func (c *Connection) processPacketFromKerepStateServerReady(p *shila.Packet) error {
	// Put packet into egress queue of connection.
	// If the connection is established at one one point, these packets
	// are sent. If not they are lost.
	// (--> Take care, could block if too many packets are in queue
	c.assignPacketHeaderFromConnectionToPacket(p)
	c.channels.NetworkEndpoint.Egress <- p
	return nil
}

func (c *Connection) processPacketFromKerepStateEstablished(p *shila.Packet) error {
	c.assignPacketHeaderFromConnectionToPacket(p)
	c.touched = time.Now()
	c.channels.NetworkEndpoint.Egress <- p
	return nil
}

func (c *Connection) processPacketFromKerepStateClosed(p *shila.Packet) error {
	log.Info.Println("Drop packet - sent through closed connection.")
	return nil
}

func (c *Connection) ID() string {
	return string(c.id)
}

func (c *Connection) createAndEstablishContactingChannel() error {

	// TODO: Make size of channels configurable
	c.channels.Contacting.Channels.Egress =  make(shila.PacketChannel, 10)
	c.channels.Contacting.Channels.Ingress = make(shila.PacketChannel, 10)
	// TODO: Maybe assign a "default" path here, since this is the connection used for the contacting
	c.channels.Contacting.Endpoint = networkEndpoint.Generator{}.NewClient(c.header.Dst, c.header.Path,
		c.channels.Contacting.Channels.Egress, c.channels.Contacting.Channels.Ingress)

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