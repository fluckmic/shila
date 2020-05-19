package connection

import (
	"shila/core/model"
	"shila/kernelSide"
	"shila/networkSide"
	"sync"
	"time"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type Mapping struct {
	kernelSide 	*kernelSide.Manager
	networkSide *networkSide.Manager
	routing 	*model.Mapping
	connections map[model.IPConnectionIdentifierKey] *Connection
	lock		sync.Mutex
}

func NewMapping(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing *model.Mapping) *Mapping {
	m := &Mapping{kernelSide, networkSide, routing,
		make(map[model.IPConnectionIdentifierKey] *Connection), sync.Mutex{}}
	go m.vacuum()
	return m
}

// Go through all connections and check whether they are still in use or not.
// If the connections were not touched for a certain time, they are removed from
// the mapping. Note that they are not delete, just set to closed and removed
// from the mapping. Deleting of the connection is done by the GC as soon as there
// is no more reference pointing to the connection.
func (m *Mapping) vacuum() {
	for {
		// TODO: Make vacuum interval configurable
		time.Sleep(time.Second)
		m.lock.Lock()
		for key, con := range m.connections {
			if time.Since(con.touched) > (20 * time.Second) {
				con.Close()
				delete(m.connections, key)
			}
		}
		m.lock.Unlock()
	}
}

func (m *Mapping) Retrieve(id model.IPConnectionIdentifierKey) *Connection {
	m.lock.Lock()
	defer m.lock.Unlock()
	if con, ok := m.connections[id]; ok {
		return con
	} else {
		newCon := New(m.kernelSide, m.networkSide, m.routing, id)
		m.connections[id] = newCon
		return newCon
	}
}