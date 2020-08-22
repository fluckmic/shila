//
package connection

import (
	"shila/config"
	"shila/core/router"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/networkSide"
	"sync"
	"time"
)

type Mapping struct {
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	routing     router.Router
	connections map[shila.TCPFlowKey] *Connection
	lock        sync.Mutex
}

func NewMapping(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing router.Router) Mapping {
	m := Mapping{
		kernelSide: 	kernelSide,
		networkSide: 	networkSide,
		routing: 		routing,
		connections: 	make(map[shila.TCPFlowKey] *Connection)}
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
		time.Sleep(time.Duration(config.Config.Connection.VacuumInterval) * time.Second)
		m.lock.Lock()
		for key, con := range m.connections {
			if time.Since(con.touched) > (time.Duration(config.Config.Connection.MaxTimeUntouched) * time.Second) {
				con.Close(shila.TolerableError("Connection got dusty."))
				delete(m.connections, key)
			}
		}
		m.lock.Unlock()
	}
}

func (m *Mapping) Retrieve(flow shila.Flow) *Connection {
	m.lock.Lock()
	defer m.lock.Unlock()
	key := flow.TCPFlow.Key()
	if con, ok := m.connections[key]; ok {
		return con
	} else {
		newCon := New(flow, m.kernelSide, m.networkSide, m.routing)
		m.connections[key] = newCon
		return newCon
	}
}

func (m *Mapping) Close(key shila.TCPFlowKey, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if con, ok := m.connections[key]; ok {
		con.Close(err)
	}
	// Cannot close a none existent connection.
}