package connection

import (
	"fmt"
	"shila/kernelSide"
	"shila/kernelSide/kernelEndpoint"
	"shila/log"
	"shila/networkSide"
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
	id          ID
	header      *shila.PacketHeader
	state       State
	channels    Channels
	lock        sync.Mutex
	touched     time.Time
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
}

type Channels struct {
	KernelEndpoint  shila.TrafficChannels 	// Kernel end point
	NetworkEndpoint shila.TrafficChannels 	// Network end point
	Contacting      shila.TrafficChannels	// End point for connection establishment
}

func New(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, id ID) *Connection {
	return &Connection{id, nil ,Raw, Channels{} ,
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

	case ClientReady:		p.SetPacketHeader(c.header)
							c.touched = time.Now()
							c.channels.Contacting.Egress <- p
							return nil

	case ServerReady:		// Put packet into egress queue of connection. If the connection is established at one one point, these packets
							// are sent. If not they are lost. (--> Take care, could block if too many packets are in queue
							p.SetPacketHeader(c.header)
							c.channels.NetworkEndpoint.Egress <- p
							return nil

	case Established: 		p.SetPacketHeader(c.header)
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
	switch c.state {

	case Raw:				return Error(fmt.Sprint("Cannot process packet - Invalid connection state ", c.state))

	case ClientReady:		return Error(fmt.Sprint("Cannot process packet - Invalid connection state ", c.state))

	case ServerReady:		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							c.state = Established
							return nil

	case Established: 		c.touched = time.Now()
							c.channels.KernelEndpoint.Egress <- p
							return nil

	case Closed: 			log.Info.Println("Drop packet - Sent through closed connection.")
							return nil

	default: 				return Error(fmt.Sprint("Cannot process packet - Unknown connection state."))
	}
}

func (c *Connection) processPacketFromKerepStateRaw(p *shila.Packet) error {

	// Assign the channels from the device through which the packet was received.
	var ep interface{} = p.EntryPoint()
	if entryPoint, ok := ep.(*kernelEndpoint.Device); ok {
		c.channels.KernelEndpoint.Ingress = entryPoint.TrafficChannels().Ingress // ingress from kernel end point
		c.channels.KernelEndpoint.Egress  = entryPoint.TrafficChannels().Egress  // egress towards kernel end point
	} else {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - Invalid entry point type."))
	}

	// Create the packet header which is associated with the connection
	if header, err := p.PacketHeader(); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	} else {
		c.header = header
	}

	// Create the contacting connection
	if channels, err := c.networkSide.EstablishNewContactingClientEndpoint(c.header.Dst); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
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
			log.Info.Print("Unable to establish traffic connection - ", err.Error())
		} else {
			c.channels.NetworkEndpoint = channels
			c.lock.Lock()
			defer c.lock.Unlock()
			c.state = Established
		}
	}()

	// Set new state
	c.state = ClientReady

	return nil
}

func (c *Connection) processPacketFromContactingEndpointStateRaw(p *shila.Packet) error {

	// Get the kernel endpoint from the kernel side manager
	if id, err := p.ID(); err == nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	} else {
		if channels, ok := c.kernelSide.GetTrafficChannels(id.Dst.IP); ok {
			c.channels.KernelEndpoint = channels
		} else {
			c.state = Closed
			return Error(fmt.Sprint("Cannot process packet - No kernel endpoint for ", id.Dst.IP))
		}
	}

	// Send packet to kernel endpoint
	// --> 	Could still be that connection cannot be established, since we have no idea if there is actually a server listening
	c.channels.KernelEndpoint.Egress <- p

	// Create the packet header which is associated with the connection
	if header, err := p.PacketHeader(); err != nil {
		c.state = Closed
		return Error(fmt.Sprint("Cannot process packet - ", err.Error()))
	} else {
		c.header = header
	}

	// Request new incoming connection from network side.
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