package networkSide

import (
	"fmt"
	"shila/config"
	"shila/core/model"
	"shila/log"
	"shila/networkSide/networkEndpoint"
	"sync"
	"time"
)

type Manager struct {
	config                    	 config.Config
	contactingServer          	 model.ServerNetworkEndpoint
	serverTrafficEndpoints    	 model.ServerEndpointMapping
	clientContactingEndpoints 	 model.ClientEndpointMapping
	clientTrafficEndpoints    	 model.ClientEndpointMapping
	isRunning                 	 bool
	workingSide               	 chan model.TrafficChannels
	lock					  	 sync.Mutex
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, workingSide chan model.TrafficChannels) *Manager {
	return &Manager{config,nil,
		nil, nil, nil,
		false, workingSide, sync.Mutex{}}
}

func (m *Manager) Setup() error {

	if m.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel side - Already setup."))
	}

	// Create the contacting server
	addr := networkEndpoint.Generator{}.NewLocalAddress(m.config.NetworkSide.ContactingServerPort)
	m.contactingServer = networkEndpoint.Generator{}.NewServer(addr, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)

	// Create the mappings
	m.serverTrafficEndpoints 		= make(model.ServerEndpointMapping)

	m.clientContactingEndpoints 	= make(model.ClientEndpointMapping)
	m.clientTrafficEndpoints 		= make(model.ClientEndpointMapping)

	return nil
}

func (m *Manager) Start() error {

	if !m.IsSetup() {
		return Error(fmt.Sprint("Cannot start network side - Network side not yet setup."))
	}

	if m.IsRunning() {
		return Error(fmt.Sprint("Cannot start network side - Network side already running."))
	}

	log.Verbose.Println("Starting network side...")

	if err := m.contactingServer.SetupAndRun(); err != nil {
		return Error(fmt.Sprint("Cannot start network side - ", err.Error()))
	}

	// Announce the traffic channels to the working side
	m.workingSide <- m.contactingServer.TrafficChannels()

	m.isRunning = true

	log.Verbose.Println("Network side started.")

	return nil
}

func (m *Manager) CleanUp() error {

	var err error = nil

	if !m.IsSetup() {
		return err
	}

	log.Info.Println("Mopping up the network side..")

	err = m.tearDownClientContactingEndpoints(); m.clearClientContactingEndpoints()
	err = m.tearDownClientTrafficEndpoints();	 m.clearClientTrafficEndpoints()
	err = m.tearDownServerTrafficEndpoints();	 m.clearServerTrafficEndpoints()

	err = m.contactingServer.TearDown(); m.contactingServer = nil

	m.isRunning 	   = false

	log.Info.Println("Network side mopped.")

	return err
}

func (m *Manager) EstablishNewServerEndpoint(addr model.NetworkAddress) (model.TrafficChannels, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	// If there already exists a server endpoint listening for addr, return its channels
	if epc, ok := m.serverTrafficEndpoints[networkEndpoint.Generator{}.GetAddressKey(addr)]; ok {
		epc.ConnectionCount++
		return epc.Endpoint.TrafficChannels(), nil
	} else {
		// Otherwise establish a new one
		newServerEndpoint := networkEndpoint.Generator{}.NewServer(addr, model.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newServerEndpoint.SetupAndRun(); err != nil {
			return model.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new server endpoint. - ", err.Error()))
		}
		// Add the server endpoint to the corresponding mapping
		m.serverTrafficEndpoints[networkEndpoint.Generator{}.GetAddressKey(addr)] =
			model.ServerNetworkEndpointAndConnectionCount{Endpoint: newServerEndpoint, ConnectionCount: 1}

		// Announce the new traffic channels to the working side
		m.workingSide <- newServerEndpoint.TrafficChannels()

		return newServerEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) EstablishNewContactingClientEndpoint(addr model.NetworkAddress) (model.TrafficChannels, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	// Fetch the default contacting path and check if there already exists
	// a contacting endpoint which should not be the case.
	path := networkEndpoint.Generator{}.GetDefaultContactingPath(addr)
	if _, ok := m.clientContactingEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)]; ok {
		return model.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new contacting client endpoint. - Endpoint already exists."))
	} else {
		// Establish a new contacting client endpoint
		newClientContactingEndpoint := networkEndpoint.Generator{}.NewClient(addr, path, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newClientContactingEndpoint.SetupAndRun(); err != nil {
			return model.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new contacting client endpoint. - ", err.Error()))
		}
		// Add it to the corresponding mapping
		m.clientContactingEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)] = newClientContactingEndpoint
		// Announce the new traffic channels to the working side
		m.workingSide <- newClientContactingEndpoint.TrafficChannels()

		return newClientContactingEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) EstablishNewTrafficClientEndpoint(addr model.NetworkAddress, path model.NetworkPath) (model.TrafficChannels, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clientTrafficEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)]; ok {
		return model.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new traffic client endpoint. - Endpoint already exists."))
	} else {
		// Otherwise establish a new one
		newClientTrafficEndpoint := networkEndpoint.Generator{}.NewClient(addr, path, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		// Wait a certain amount of time to give the server endpoint time to establish itself
		time.Sleep(time.Duration(m.config.NetworkSide.WaitingTimeTrafficConnEstablishment) * time.Second)
		if err := newClientTrafficEndpoint.SetupAndRun(); err != nil {
			return model.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new traffic client endpoint. - ", err.Error()))
		}
		// Add it to the corresponding mapping
		m.clientTrafficEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)] = newClientTrafficEndpoint
		// Announce the new traffic channels to the working side
		m.workingSide <- newClientTrafficEndpoint.TrafficChannels()

		// The removal of the corresponding client contacting endpoint is triggered by the connection
		// itself after obtaining the lock to change its state to established. Otherwise we are in danger
		// that we close the endpoint to early.

		return newClientTrafficEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) TeardownTrafficSeverEndpoint(addr model.NetworkAddress) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	key := networkEndpoint.Generator{}.GetAddressKey(addr)
	if epc, ok := m.serverTrafficEndpoints[key]; ok {
		epc.ConnectionCount--
		if epc.ConnectionCount == 0 {
			err := epc.Endpoint.TearDown()
			delete(m.serverTrafficEndpoints, key)
			return err
		}
	}
	return nil
}

func (m *Manager) TeardownContactingClientEndpoint(addr model.NetworkAddress) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	path := networkEndpoint.Generator{}.GetDefaultContactingPath(addr)
	key  := networkEndpoint.Generator{}.GetAddressPathKey(addr, path)
	if ep, ok := m.clientContactingEndpoints[key]; ok {
		err := ep.TearDown()
		delete(m.clientTrafficEndpoints, key)
		return err
	}
	return nil
}

func (m *Manager) TeardownTrafficClientEndpoint(addr model.NetworkAddress, path model.NetworkPath) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	key := networkEndpoint.Generator{}.GetAddressPathKey(addr, path)
	if ep, ok := m.clientTrafficEndpoints[key]; ok {
		err := ep.TearDown()
		delete(m.clientTrafficEndpoints, key)
		return err
	}
	return nil
}

func (m *Manager) IsSetup() bool {
	return m.contactingServer != nil
}

func (m *Manager) IsRunning() bool {
	return m.isRunning
}

func (m *Manager) clearServerTrafficEndpoints() {
	for k := range m.serverTrafficEndpoints {
		delete(m.serverTrafficEndpoints, k)
	}
}

func (m *Manager) clearClientContactingEndpoints() {
	for k := range m.clientContactingEndpoints {
		delete(m.clientContactingEndpoints, k)
	}
}

func (m *Manager) clearClientTrafficEndpoints() {
	for k := range m.clientTrafficEndpoints {
		delete(m.clientTrafficEndpoints, k)
	}
}

func (m *Manager) tearDownServerTrafficEndpoints() error {
	var err error = nil
	for _, epc := range m.serverTrafficEndpoints {
		err = epc.Endpoint.TearDown()
	}
	return err
}

func (m *Manager) tearDownClientContactingEndpoints() error {
	var err error = nil
	for _, ep := range m.clientContactingEndpoints {
		err = ep.TearDown()
	}
	return err
}

func (m *Manager) tearDownClientTrafficEndpoints() error {
	var err error = nil
	for _, ep := range m.clientTrafficEndpoints {
		err = ep.TearDown()
	}
	return err
}