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
	workingSide               	 chan model.PacketChannelAnnouncement
	lock					  	 sync.Mutex
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, workingSide chan model.PacketChannelAnnouncement) *Manager {
	return &Manager{config,nil,
		nil, nil, nil,
		false, workingSide, sync.Mutex{}}
}

func (m *Manager) Setup() error {

	if m.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel side - Already setup."))
	}

	// Create the contacting server
	netConnId := model.NetworkConnectionIdentifier{Src: networkEndpoint.Generator{}.NewLocalAddress(m.config.NetworkSide.ContactingServerPort)}
	m.contactingServer = networkEndpoint.Generator{}.NewServer(netConnId, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)

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
	m.workingSide <- model.PacketChannelAnnouncement{Announcer: m.contactingServer, Channel: m.contactingServer.TrafficChannels().Ingress}

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

func (m *Manager) EstablishNewServerEndpoint(netConnId model.NetworkConnectionIdentifier) (model.PacketChannels, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	// If there already exists a server endpoint listening for addr, return its channels
	key := model.KeyGenerator{}.NetworkAddressKey(netConnId.Src)
	if epc, ok := m.serverTrafficEndpoints[key]; ok {
		epc.ConnectionCount++
		return epc.Endpoint.TrafficChannels(), nil
	} else {
		// Otherwise establish a new one
		newServerEndpoint := networkEndpoint.Generator{}.NewServer(netConnId, model.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newServerEndpoint.SetupAndRun(); err != nil {
			return model.PacketChannels{}, Error(fmt.Sprint("Unable to establish new server endpoint. - ", err.Error()))
		}
		// Add the server endpoint to the corresponding mapping
		m.serverTrafficEndpoints[key] =
			model.ServerNetworkEndpointAndConnectionCount{Endpoint: newServerEndpoint, ConnectionCount: 1}

		// Announce the new traffic channels to the working side
		m.workingSide <- model.PacketChannelAnnouncement{Announcer: newServerEndpoint, Channel: newServerEndpoint.TrafficChannels().Ingress}

		return newServerEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) EstablishNewContactingClientEndpoint(netConnId model.NetworkConnectionIdentifier) (model.NetworkConnectionIdentifier, model.PacketChannels, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	path := networkEndpoint.Generator{}.GetDefaultContactingPath(netConnId.Dst)
	key  := model.KeyGenerator{}.NetworkAddressAndPathKey(netConnId.Dst, path)

	// Fetch the default contacting contactingPath and check if there already exists
	// a contacting endpoint which should not be the case.
	if _, ok := m.clientContactingEndpoints[key]; ok {
		return model.NetworkConnectionIdentifier{}, model.PacketChannels{}, Error(fmt.Sprint("Unable to establish new client {",
			model.ContactingNetworkEndpoint, "} connected to {", netConnId.Dst, "}. - Endpoint already exists."))
	} else {
		// Establish a new contacting client endpoint
		newContactingClientEndpoint := networkEndpoint.Generator{}.NewClient(netConnId, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		if contactingNetConnId, err := newContactingClientEndpoint.SetupAndRun(); err != nil {
			return model.NetworkConnectionIdentifier{}, model.PacketChannels{}, Error(fmt.Sprint("Unable to establish new client {",
				model.ContactingNetworkEndpoint, "} connected to {", netConnId.Dst, "}. - ", err.Error()))
		} else {
			// Add it to the corresponding mapping
			m.clientContactingEndpoints[key] = newContactingClientEndpoint
			// Announce the new traffic channels to the working side
			m.workingSide <- model.PacketChannelAnnouncement{Announcer: newContactingClientEndpoint, Channel: newContactingClientEndpoint.TrafficChannels().Ingress}

			return contactingNetConnId, newContactingClientEndpoint.TrafficChannels(), nil
		}
	}
}

func (m *Manager) EstablishNewTrafficClientEndpoint(netConnId model.NetworkConnectionIdentifier) (model.NetworkConnectionIdentifier, model.PacketChannels, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	key := model.KeyGenerator{}.NetworkAddressAndPathKey(netConnId.Dst, netConnId.Path)
	if _, ok := m.clientTrafficEndpoints[key]; ok {
		return model.NetworkConnectionIdentifier{}, model.PacketChannels{},
		Error(fmt.Sprint("Unable to establish new traffic client endpoint. - Endpoint already exists."))
	} else {
		// Otherwise establish a new one
		newTrafficClientEndpoint := networkEndpoint.Generator{}.NewClient(netConnId, model.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		// Wait a certain amount of time to give the server endpoint time to establish itself
		time.Sleep(time.Duration(m.config.NetworkSide.WaitingTimeTrafficConnEstablishment) * time.Second)
		if trafficNetConnId, err := newTrafficClientEndpoint.SetupAndRun(); err != nil {
			return model.NetworkConnectionIdentifier{}, model.PacketChannels{},
			Error(fmt.Sprint("Unable to establish new traffic client endpoint. - ", err.Error()))
		} else {
			// Add it to the corresponding mapping
			m.clientTrafficEndpoints[key] = newTrafficClientEndpoint
			// Announce the new traffic channels to the working side
			m.workingSide <- model.PacketChannelAnnouncement{Announcer: newTrafficClientEndpoint, Channel: newTrafficClientEndpoint.TrafficChannels().Ingress}

			// The removal of the corresponding client contacting endpoint is triggered by the connTr
			// itself after obtaining the lock to change its state to established. Otherwise we are in danger
			// that we close the endpoint to early.

			return trafficNetConnId, newTrafficClientEndpoint.TrafficChannels(), nil
			}
		}
}

func (m *Manager) TeardownTrafficSeverEndpoint(addr model.NetworkAddress) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	key :=  model.KeyGenerator{}.NetworkAddressKey(addr)
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
	key  := model.KeyGenerator{}.NetworkAddressAndPathKey(addr, path)
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

	key := model.KeyGenerator{}.NetworkAddressAndPathKey(addr, path)
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