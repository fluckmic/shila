package connection

import (
	"fmt"
	"sync"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type ID string

type Mapping struct {
	connections map[ID] *Connection
	lock		sync.Mutex
}

func NewMapping() *Mapping {
	return &Mapping{make(map[ID] *Connection), sync.Mutex{}}
}

func (m *Mapping) Retrieve(id ID) *Connection {
	m.lock.Lock()
	defer m.lock.Unlock()
	if con, ok := m.connections[id]; ok {
		return con
	} else {
		newCon := New()
		m.connections[id] = newCon
		return newCon
	}
}

func (m *Mapping) add(id ID, con *Connection) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.connections[id]; !ok {
		m.connections[id] = con
		return nil
	} else {
		return Error(fmt.Sprint("Connection with id ", id, " already present."))
	}
}

func (m *Mapping) remove(id ID, con *Connection) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.connections[id]; ok {
		delete(m.connections, id)
		return nil
	} else {
		return Error(fmt.Sprint("Non connection with id ", id, " present."))
	}
}