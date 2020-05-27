package connection

import (
	"shila/core/netflow"
	"shila/core/shila"
	"shila/kernelSide"
	"shila/networkSide"
	"sync"
	"time"
)

type Mapping struct {
	kernelSide 	*kernelSide.Manager
	networkSide *networkSide.Manager
	routing 	*netflow.Router
	connections map[shila.IPFlowKey] *Connection
	lock		sync.Mutex
}

func NewMapping(kernelSide *kernelSide.Manager, networkSide *networkSide.Manager, routing *netflow.Router) Mapping {
	m := Mapping{
		kernelSide: 	kernelSide,
		networkSide: 	networkSide,
		routing: 		routing,
		connections: 	make(map[shila.IPFlowKey] *Connection)}
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
		time.Sleep(Config.VacuumInterval)
		m.lock.Lock()
		for key, con := range m.connections {
			if time.Since(con.touched) > (Config.MaxTimeUntouched) {
				con.Close(shila.ThirdPartyError("Connection got dusty."))
				delete(m.connections, key)
			}
		}
		m.lock.Unlock()
	}
}

func (m *Mapping) Retrieve(flow shila.Flow) *Connection {
	m.lock.Lock()
	defer m.lock.Unlock()
	key := flow.IPFlow.Key()
	if con, ok := m.connections[key]; ok {
		return con
	} else {
		newCon := New(flow, m.kernelSide, m.networkSide, m.routing)
		m.connections[key] = newCon
		return newCon
	}
}