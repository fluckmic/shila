package connection

import (
	"shila/kersi/kerep"
	"shila/network"
	"shila/shila"
	"sync"
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
	State State
	Kerep *kerep.Device
	Newep *network.Endpoint
	lock  sync.Mutex
}

func New() *Connection {
	return &Connection{Raw, nil, nil, sync.Mutex{}}
}

func (c *Connection) processPacket(packet *shila.Packet) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return nil
}