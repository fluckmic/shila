package networkSide

import (
	"fmt"
	"shila/config"
	"shila/core/shila"
	"shila/log"
	"shila/networkSide/networkEndpoint"
	"sync"
	"time"
)

type Manager struct {
	config                    config.Config
	contactingServer          shila.ServerNetworkEndpoint
	serverTrafficEndpoints    shila.ServerEndpointMapping
	clientContactingEndpoints shila.ClientEndpointMapping
	clientTrafficEndpoints    shila.ClientEndpointMapping
	workingSide               chan shila.PacketChannelAnnouncement
	lock                      sync.Mutex
	state                     shila.EntityState
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, workingSide chan shila.PacketChannelAnnouncement) *Manager {
	return &Manager{
		config:      config,
		workingSide: workingSide,
		lock:        sync.Mutex{},
		state:       shila.EntityState{EntityStateIdentifier: shila.Uninitialized},
	}
}

func (m *Manager) Setup() error {

	if m.state.Get() != shila.Uninitialized {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	// Create the contacting server
	netConnId := shila.NetFlow{Src: networkEndpoint.Generator{}.NewLocalAddress(m.config.NetworkSide.ContactingServerPort)}
	m.contactingServer = networkEndpoint.Generator{}.NewServer(netConnId, shila.ContactingNetworkEndpoint, m.config.NetworkEndpoint)

	// Create the mappings
	m.serverTrafficEndpoints 		= make(shila.ServerEndpointMapping)

	m.clientContactingEndpoints 	= make(shila.ClientEndpointMapping)
	m.clientTrafficEndpoints 		= make(shila.ClientEndpointMapping)

	m.state.Set(shila.Initialized)

	return nil
}

func (m *Manager) Start() error {

	if m.state.Get() != shila.Initialized {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	//log.Verbose.Println("Starting network side...")

	if err := m.contactingServer.SetupAndRun(); err != nil {
		return Error(fmt.Sprint("Unable to setup and run contacting server. - ", err.Error()))
	}

	// Announce the traffic channels to the working side
	m.workingSide <- shila.PacketChannelAnnouncement{Announcer: m.contactingServer, Channel: m.contactingServer.TrafficChannels().Ingress}

	m.state.Set(shila.Running)

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

	m.state.Set(shila.TornDown)

	log.Info.Println("Network side torn down.")

	return err
}

func (m *Manager) EstablishNewTrafficServerEndpoint(flow shila.Flow) (shila.PacketChannels, error) {

	if m.state.Get() != shila.Running {
		return shila.PacketChannels{}, Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key     := shila.GetNetworkAddressKey(flow.NetFlow.Src)
	sep, ok := m.serverTrafficEndpoints[key]

	// If there is no server endpoint listening, we first have to set one up.
	if !ok {
		newServerEndpoint   := networkEndpoint.Generator{}.NewServer(flow.NetFlow, shila.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newServerEndpoint.SetupAndRun(); err != nil {
			return shila.PacketChannels{}, Error(fmt.Sprint("Unable to setup and run new server {",
				shila.ContactingNetworkEndpoint, "} listening to {", flow.NetFlow.Src, "}. - ", err.Error()))
		}
		// Add the endpoint to the mapping
		sep = shila.ServerNetworkEndpointMapping{ServerNetworkEndpoint: newServerEndpoint, IPConnectionMapping: make(shila.IPConnectionMapping)}
		m.serverTrafficEndpoints[key] = sep

		// Announce the new traffic channels to the working side
		m.workingSide <- shila.PacketChannelAnnouncement{Announcer: newServerEndpoint, Channel: newServerEndpoint.TrafficChannels().Ingress}
	}

	sep.AddIPFlowKey(flow.IPFlow.Key())

	return sep.TrafficChannels(), nil
}

func (m *Manager) EstablishNewContactingClientEndpoint(flow shila.Flow) (shila.NetFlow, shila.PacketChannels, error) {

	if m.state.Get() != shila.Running {
		return shila.NetFlow{}, shila.PacketChannels{},
			Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	flow.NetFlow.Path = networkEndpoint.Generator{}.GetDefaultContactingPath(flow.NetFlow.Dst)
	flow.NetFlow.Dst  = networkEndpoint.Generator{}.GenerateContactingAddress(flow.NetFlow.Dst)

	// Fetch the default contacting contactingPath and check if there already exists
	// a contacting endpoint which should not be the case.
	if _, ok := m.clientContactingEndpoints[flow.IPFlow.Key()]; ok {
		return shila.NetFlow{}, shila.PacketChannels{},
		Error(fmt.Sprint("Endpoint {", flow.IPFlow.Key(), "} already exists."))
	} else {
		// Establish a new contacting client endpoint
		newContactingClientEndpoint := networkEndpoint.Generator{}.NewClient(flow.NetFlow, shila.ContactingNetworkEndpoint, m.config.NetworkEndpoint)
		if contactingNetConnId, err := newContactingClientEndpoint.SetupAndRun(); err != nil {
			return shila.NetFlow{}, shila.PacketChannels{}, Error(fmt.Sprint("Unable to setup and run new client {",
				shila.ContactingNetworkEndpoint, "} connected to {", flow.NetFlow.Dst, "}. - ", err.Error()))
		} else {
			// Add it to the corresponding mapping
			m.clientContactingEndpoints[flow.IPFlow.Key()] = newContactingClientEndpoint
			// Announce the new traffic channels to the working side
			m.workingSide <- shila.PacketChannelAnnouncement{Announcer: newContactingClientEndpoint, Channel: newContactingClientEndpoint.TrafficChannels().Ingress}

			return contactingNetConnId, newContactingClientEndpoint.TrafficChannels(), nil
		}
	}
}

func (m *Manager) EstablishNewTrafficClientEndpoint(flow shila.Flow) (shila.NetFlow, shila.PacketChannels, error) {

	if m.state.Get() != shila.Running {
		return shila.NetFlow{}, shila.PacketChannels{},
			Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clientTrafficEndpoints[flow.IPFlow.Key()]; ok {
		return shila.NetFlow{}, shila.PacketChannels{},
			Error(fmt.Sprint("Endpoint {", flow.IPFlow.Key(), "} already exists."))
	} else {
		// Otherwise establish a new one
		newTrafficClientEndpoint := networkEndpoint.Generator{}.NewClient(flow.NetFlow, shila.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		// Wait a certain amount of time to give the server endpoint time to establish itself
		time.Sleep(time.Duration(m.config.NetworkSide.WaitingTimeTrafficConnEstablishment) * time.Second)
		if trafficNetConnId, err := newTrafficClientEndpoint.SetupAndRun(); err != nil {
			return shila.NetFlow{}, shila.PacketChannels{},
			Error(fmt.Sprint("Unable to establish new traffic client endpoint. - ", err.Error()))
		} else {
			// Add it to the corresponding mapping
			m.clientTrafficEndpoints[flow.IPFlow.Key()] = newTrafficClientEndpoint
			// Announce the new traffic channels to the working side
			m.workingSide <- shila.PacketChannelAnnouncement{Announcer: newTrafficClientEndpoint, Channel: newTrafficClientEndpoint.TrafficChannels().Ingress}

			// The removal of the corresponding client contacting endpoint is triggered by the connTr
			// itself after obtaining the lock to change its state to established. Otherwise we are in danger
			// that we close the endpoint to early.

			return trafficNetConnId, newTrafficClientEndpoint.TrafficChannels(), nil
			}
		}
}

func (m *Manager) TeardownTrafficSeverEndpoint(flow shila.Flow) error {

	if m.state.Get() != shila.Running {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key     := shila.GetNetworkAddressKey(flow.NetFlow.Src)
	if ep, ok := m.serverTrafficEndpoints[key]; ok {

		ep.RemoveIPFlowKey(flow.IPFlow.Key())
		if ep.Empty() {
			err := ep.TearDown()
			delete(m.clientContactingEndpoints, flow.IPFlow.Key())
			return err
		}

	}

	return nil
}

func (m *Manager) TeardownContactingClientEndpoint(ipFlow shila.IPFlow) error {

	if m.state.Get() != shila.Running {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if ep, ok := m.clientContactingEndpoints[ipFlow.Key()]; ok {
		err := ep.TearDown()
		delete(m.clientContactingEndpoints, ipFlow.Key())
		return err
	}

	return nil
}

func (m *Manager) TeardownTrafficClientEndpoint(ipFlow shila.IPFlow) error {

	if m.state.Get() != shila.Running {
		return Error(fmt.Sprint("Entity in wrong state {", m.state.Get(), "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if ep, ok := m.clientTrafficEndpoints[ipFlow.Key()]; ok {
		err := ep.TearDown()
		delete(m.clientTrafficEndpoints, ipFlow.Key())
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