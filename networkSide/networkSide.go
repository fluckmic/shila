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
	workingSide               	 chan model.PacketChannelAnnouncement
	lock					  	 sync.Mutex
	state						 model.EntityState
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, workingSide chan model.PacketChannelAnnouncement) *Manager {
	return &Manager{
		config: 	 config,
		workingSide: workingSide,
		lock: 		 sync.Mutex{},
		state: 		 model.EntityState{EntityStateIdentifier: model.Uninitialized},
	}
}

func (m *Manager) Setup() error {

	if m.state.Get() != model.Uninitialized {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	// Create the contacting server
	netConnId := model.NetworkConnectionIdentifier{Src: networkEndpoint.Generator{}.NewLocalAddress(m.config.NetworkSide.ContactingServerPort)}
	m.contactingServer = networkEndpoint.Generator{}.NewServer(netConnId, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)

	// Create the mappings
	m.serverTrafficEndpoints 		= make(model.ServerEndpointMapping)

	m.clientContactingEndpoints 	= make(model.ClientEndpointMapping)
	m.clientTrafficEndpoints 		= make(model.ClientEndpointMapping)

	m.state.Set(model.Initialized)

	return nil
}

func (m *Manager) Start() error {

	if m.state.Get() != model.Initialized {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	//log.Verbose.Println("Starting network side...")

	if err := m.contactingServer.SetupAndRun(); err != nil {
		return Error(fmt.Sprint("Unable to setup and run contacting server. - ", err.Error()))
	}

	// Announce the traffic channels to the working side
	m.workingSide <- model.PacketChannelAnnouncement{Announcer: m.contactingServer, Channel: m.contactingServer.TrafficChannels().Ingress}

	m.state.Set(model.Running)

	//log.Verbose.Println("Network side started.")

	return nil
}

func (m *Manager) CleanUp() error {

	var err error = nil

	log.Info.Println("Tear down the network side..")

	err = m.tearDownAndRemoveClientContactingEndpoints()
	err = m.tearDownAndRemoveClientTrafficEndpoints()
	err = m.tearDownAndRemoveServerTrafficEndpoints()

	err = m.contactingServer.TearDown()
	m.contactingServer = nil

	m.state.Set(model.TornDown)

	log.Info.Println("Network side torn down.")

	return err
}

func (m *Manager) EstablishNewTrafficServerEndpoint(IPConnIdKey model.IPConnectionIdentifierKey, netConnId model.NetworkConnectionIdentifier) (model.PacketChannels, error) {

	if m.state.Get() != model.Running {
		return model.PacketChannels{}, Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key     := model.KeyGenerator{}.NetworkAddressKey(netConnId.Src)
	sep, ok := m.serverTrafficEndpoints[key]

	// If there is no server endpoint listening, we first have to set one up.
	if !ok {
		newServerEndpoint   := networkEndpoint.Generator{}.NewServer(netConnId, model.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newServerEndpoint.SetupAndRun(); err != nil {
			return model.PacketChannels{}, Error(fmt.Sprint("Unable to setup and run new server {",
				model.ContactingNetworkEndpoint, "} listening to {", netConnId.Src, "}. - ", err.Error()))
		}
		// Add the endpoint to the mapping
		sep := model.ServerNetworkEndpointMapping{ServerNetworkEndpoint: newServerEndpoint}
		m.serverTrafficEndpoints[key] = sep

		// Announce the new traffic channels to the working side
		m.workingSide <- model.PacketChannelAnnouncement{Announcer: newServerEndpoint, Channel: newServerEndpoint.TrafficChannels().Ingress}
	}

	sep.AddIPConnectionIdentifierKey(IPConnIdKey)

	return sep.TrafficChannels(), nil
}

func (m *Manager) EstablishNewContactingClientEndpoint(IPConnIdKey model.IPConnectionIdentifierKey, netConnId model.NetworkConnectionIdentifier) (model.NetworkConnectionIdentifier, model.PacketChannels, error) {

	if m.state.Get() != model.Running {
		return model.NetworkConnectionIdentifier{}, model.PacketChannels{},
			Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	netConnId.Path = networkEndpoint.Generator{}.GetDefaultContactingPath(netConnId.Dst)
	netConnId.Dst  = networkEndpoint.Generator{}.GenerateContactingAddress(netConnId.Dst)

	// Fetch the default contacting contactingPath and check if there already exists
	// a contacting endpoint which should not be the case.
	if _, ok := m.clientContactingEndpoints[IPConnIdKey]; ok {
		return model.NetworkConnectionIdentifier{}, model.PacketChannels{},
		Error(fmt.Sprint("Endpoint {", IPConnIdKey, "} already exists."))
	} else {
		// Establish a new contacting client endpoint
		newContactingClientEndpoint := networkEndpoint.Generator{}.NewClient(netConnId, model.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		if contactingNetConnId, err := newContactingClientEndpoint.SetupAndRun(); err != nil {
			return model.NetworkConnectionIdentifier{}, model.PacketChannels{}, Error(fmt.Sprint("Unable to setup and run new client {",
				model.ContactingNetworkEndpoint, "} connected to {", netConnId.Dst, "}. - ", err.Error()))
		} else {
			// Add it to the corresponding mapping
			m.clientContactingEndpoints[IPConnIdKey] = newContactingClientEndpoint
			// Announce the new traffic channels to the working side
			m.workingSide <- model.PacketChannelAnnouncement{Announcer: newContactingClientEndpoint, Channel: newContactingClientEndpoint.TrafficChannels().Ingress}

			return contactingNetConnId, newContactingClientEndpoint.TrafficChannels(), nil
		}
	}
}

func (m *Manager) EstablishNewTrafficClientEndpoint(IPConnIdKey model.IPConnectionIdentifierKey, netConnId model.NetworkConnectionIdentifier) (model.NetworkConnectionIdentifier, model.PacketChannels, error) {

	if m.state.Get() != model.Running {
		return model.NetworkConnectionIdentifier{}, model.PacketChannels{},
			Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clientTrafficEndpoints[IPConnIdKey]; ok {
		return model.NetworkConnectionIdentifier{}, model.PacketChannels{},
			Error(fmt.Sprint("Endpoint {", IPConnIdKey, "} already exists."))
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
			m.clientTrafficEndpoints[IPConnIdKey] = newTrafficClientEndpoint
			// Announce the new traffic channels to the working side
			m.workingSide <- model.PacketChannelAnnouncement{Announcer: newTrafficClientEndpoint, Channel: newTrafficClientEndpoint.TrafficChannels().Ingress}

			// The removal of the corresponding client contacting endpoint is triggered by the connTr
			// itself after obtaining the lock to change its state to established. Otherwise we are in danger
			// that we close the endpoint to early.

			return trafficNetConnId, newTrafficClientEndpoint.TrafficChannels(), nil
			}
		}
}

func (m *Manager) TeardownTrafficSeverEndpoint(IPConnIdKey model.IPConnectionIdentifierKey, netConnId model.NetworkConnectionIdentifier) error {

	if m.state.Get() != model.Running {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key     := model.KeyGenerator{}.NetworkAddressKey(netConnId.Src)
	if ep, ok := m.serverTrafficEndpoints[key]; ok {

		ep.RemoveIPConnectionIdentifierKey(IPConnIdKey)
		if ep.Empty() {
			err := ep.TearDown()
			delete(m.clientContactingEndpoints, IPConnIdKey)
			return err
		}

	}

	return nil
}

func (m *Manager) TeardownContactingClientEndpoint(IPConnIdKey model.IPConnectionIdentifierKey) error {

	if m.state.Get() != model.Running {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if ep, ok := m.clientContactingEndpoints[IPConnIdKey]; ok {
		err := ep.TearDown()
		delete(m.clientContactingEndpoints, IPConnIdKey)
		return err
	}

	return nil
}

func (m *Manager) TeardownTrafficClientEndpoint(IPConnIdKey model.IPConnectionIdentifierKey) error {

	if m.state.Get() != model.Running {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if ep, ok := m.clientTrafficEndpoints[IPConnIdKey]; ok {
		err := ep.TearDown()
		delete(m.clientTrafficEndpoints, IPConnIdKey)
		return err
	}
	return nil
}

func (m *Manager) tearDownAndRemoveServerTrafficEndpoints() error {
	var err error = nil
	for key, ep := range m.serverTrafficEndpoints {
		err = ep.TearDown()
		delete(m.serverTrafficEndpoints, key)
	}
	return err
}

func (m *Manager) tearDownAndRemoveClientContactingEndpoints() error {
	var err error = nil
	for key, ep := range m.clientContactingEndpoints {
		err = ep.TearDown()
		delete(m.clientContactingEndpoints, key)
	}
	return err
}

func (m *Manager) tearDownAndRemoveClientTrafficEndpoints() error {
	var err error = nil
	for key, ep := range m.clientTrafficEndpoints {
		err = ep.TearDown()
		delete(m.clientTrafficEndpoints, key)
	}
	return err
}