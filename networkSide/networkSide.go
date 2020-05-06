package networkSide

import (
	"fmt"
	"shila/config"
	"shila/log"
	"shila/networkSide/networkEndpoint"
	"shila/shila"
	"time"
)

type Manager struct {
	config                 		config.Config
	contactingServer       		shila.ServerNetworkEndpoint
	serverTrafficEndpoints 		shila.ServerEndpointMapping
	clientContactingEndpoints 	shila.ClientEndpointMapping
	clientTrafficEndpoints 		shila.ClientEndpointMapping
	isRunning              		bool
	workingSide 				chan shila.TrafficChannels
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, workingSide chan shila.TrafficChannels) *Manager {
	return &Manager{config,nil, nil,nil,
		nil, false, workingSide}
}

func (m *Manager) Setup() error {

	if m.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel side - Already setup."))
	}

	// Create the contacting server
	addr := networkEndpoint.Generator{}.NewLocalAddress(m.config.NetworkSide.ContactingServerPort)
	m.contactingServer = networkEndpoint.Generator{}.NewServer(addr, shila.ContactingNetworkEndpoint, m.config.NetworkEndpoint)

	// Create the mappings
	m.serverTrafficEndpoints 	= make(shila.ServerEndpointMapping)
	m.clientContactingEndpoints = make(shila.ClientEndpointMapping)
	m.clientTrafficEndpoints 	= make(shila.ClientEndpointMapping)

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

func (m *Manager) CleanUp() {

}

func (m *Manager) EstablishNewServerEndpoint(addr shila.NetworkAddress) (shila.TrafficChannels, error) {

	// If there already exists a server endpoint listening for addr, return its channels
	if ep, ok := m.serverTrafficEndpoints[networkEndpoint.Generator{}.GetAddressKey(addr)]; ok {
		return ep.TrafficChannels(), nil
	} else {
		// Otherwise establish a new one
		newServerEndpoint := networkEndpoint.Generator{}.NewServer(addr, shila.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newServerEndpoint.SetupAndRun(); err != nil {
			return shila.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new server endpoint. - ", err.Error()))
		}
		// Add the server endpoint to the corresponding mapping
		m.serverTrafficEndpoints[networkEndpoint.Generator{}.GetAddressKey(addr)] = newServerEndpoint
		// Announce the new traffic channels to the working side
		m.workingSide <- newServerEndpoint.TrafficChannels()

		return newServerEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) EstablishNewContactingClientEndpoint(addr shila.NetworkAddress) (shila.TrafficChannels, error) {

	// Fetch the default contacting path and check if there already exists
	// a contacting endpoint which should not be the case.
	path := networkEndpoint.Generator{}.GetDefaultContactingPath(addr)
	if _, ok := m.clientContactingEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)]; ok {
		return shila.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new contacting client endpoint. - Endpoint already exists."))
	} else {
		// Establish a new contacting client endpoint
		newClientContactingEndpoint := networkEndpoint.Generator{}.NewClient(addr, path, shila.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newClientContactingEndpoint.SetupAndRun(); err != nil {
			return shila.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new contacting client endpoint. - ", err.Error()))
		}
		// Add it to the corresponding mapping
		m.clientContactingEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)] = newClientContactingEndpoint
		// Announce the new traffic channels to the working side
		m.workingSide <- newClientContactingEndpoint.TrafficChannels()

		return newClientContactingEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) EstablishNewTrafficClientEndpoint(addr shila.NetworkAddress, path shila.NetworkPath) (shila.TrafficChannels, error) {

	if _, ok := m.clientTrafficEndpoints[networkEndpoint.Generator{}.GetAddressPathKey(addr, path)]; ok {
		return shila.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new traffic client endpoint. - Endpoint already exists."))
	} else {
		// Otherwise establish a new one
		newClientTrafficEndpoint := networkEndpoint.Generator{}.NewClient(addr, path, shila.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		// Wait a certain amount of time to give the server endpoint time to establish itself
		time.Sleep(time.Duration(m.config.NetworkSide.WaitingTimeUntilTrafficConnectionEstablishment) * time.Second)
		if err := newClientTrafficEndpoint.SetupAndRun(); err != nil {
			return shila.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new traffic client endpoint. - ", err.Error()))
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

func (m *Manager) TeardownSeverEndpoint() {}

func (m *Manager) TeardownClientEndpoint() {}

func (m *Manager) IsSetup() bool {
	return m.contactingServer != nil
}

func (m *Manager) IsRunning() bool {
	return m.isRunning
}
