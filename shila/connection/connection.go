package connection

import (
	"fmt"
	"shila/kersi/kerep"
	"shila/log"
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
	id		ID
	state 	State
	Kerep 	*kerep.Device
	Newep 	*shila.Endpoint
	lock  	sync.Mutex
	touched time.Time
}

func New(id ID) *Connection {
	return &Connection{id, Raw, nil, nil, sync.Mutex{}, time.Now()}
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

	// Assign the entry device to the connection.
	var ep interface{} = p.EntryPoint()
	if entryPoint, ok := ep.(*kerep.Device); ok {
		c.Kerep = entryPoint	// kernel side
	} else {
		return Error(fmt.Sprint("Cannot process packet - ID mismatch: "))
	}

	c.touched = time.Now()
	return nil

	// Determine address and path

	// Send the packet via establish channel

	// Set new state
	c.state = ClientReady

	// Initiate try connect to address via path
		// Stop and clean up connection try when connection state is Closed
		// Reset timer and set state to Established when connection try succeeds

	return nil
}

func (c *Connection) processPacketFromKerepStateClientReady(p *shila.Packet) error {

	c.touched = time.Now()
	return nil

	// Send via establish channel

	return nil
}

func (c *Connection) processPacketFromKerepStateServerReady(p *shila.Packet) error {

	return nil

	// Put packet into egress queue of connection.
	// If the connection is established at one one point, these packets
	// are sent. If not they are lost. (--> Take care, could block if
	// too many packets are in queue
}

func (c *Connection) processPacketFromKerepStateEstablished(p *shila.Packet) error {

	c.touched = time.Now()
	return nil

	// Send via connection

	return nil
}

func (c *Connection) processPacketFromKerepStateClosed(p *shila.Packet) error {
	log.Info.Println("Drop packet - sent through closed connection.")
	return nil
}

func (c *Connection) ID() string {
	return string(c.id)
}