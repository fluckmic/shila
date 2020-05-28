package networkSide

import (
	"fmt"
	"shila/core/shila"
	"sync"
	"time"
)

// This part of the network is independent of the backbone protocol chosen.

type Manager struct {
	specificManager			  SpecificManager
	contactingServer          shila.NetworkServerEndpoint
	serverTrafficEndpoints    shila.MappingNetworkServerEndpoint
	clientContactingEndpoints shila.MappingNetworkClientEndpoint
	clientTrafficEndpoints    shila.MappingNetworkClientEndpoint
	workingSide               chan shila.PacketChannelAnnouncement
	lock                      sync.Mutex
	state                     shila.EntityState
}

func New(workingSide chan shila.PacketChannelAnnouncement) *Manager {
	return &Manager{
		specificManager: NewSpecificManager(),
		workingSide: 	 workingSide,
		lock:        	 sync.Mutex{},
		state:       	 shila.NewEntityState(),
	}
}

func (m *Manager) Setup() error {

	if m.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	// Create the contacting server
	localContactingNetFlow := m.specificManager.LocalContactingNetFlow()
	m.contactingServer 	   = m.specificManager.NewServer(localContactingNetFlow, shila.ContactingNetworkEndpoint)

	// Create the mappings
	m.serverTrafficEndpoints 		= make(shila.MappingNetworkServerEndpoint)
	m.clientContactingEndpoints 	= make(shila.MappingNetworkClientEndpoint)
	m.clientTrafficEndpoints 		= make(shila.MappingNetworkClientEndpoint)

	m.state.Set(shila.Initialized)

	return nil
}

func (m *Manager) Start() error {

	if m.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	if err := m.contactingServer.SetupAndRun(); err != nil {
		return shila.PrependError(err, "Unable to establish contacting server.")
	}

	// Announce the traffic channels to the working side
	m.workingSide <- shila.PacketChannelAnnouncement{Announcer: m.contactingServer, Channel: m.contactingServer.TrafficChannels().Ingress}

	m.state.Set(shila.Running)

	return nil
}

func (m *Manager) CleanUp() error {

	var err error = nil

	err = m.tearDownAndRemoveClientContactingEndpoints()
	err = m.tearDownAndRemoveClientTrafficEndpoints()
	err = m.tearDownAndRemoveServerTrafficEndpoints()

	err = m.contactingServer.TearDown()
	m.contactingServer = nil

	m.state.Set(shila.TornDown)

	return err
}

func (m *Manager) EstablishNewTrafficServerEndpoint(flow shila.Flow) (channels shila.PacketChannels, error error) {

	channels 			= shila.PacketChannels{}
	error    			= nil

	if m.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}.")); return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key     	 := shila.GetNetworkAddressKey(flow.NetFlow.Src)
	endpoint, ok := m.serverTrafficEndpoints[key]

	if ok {
		endpoint.Register(flow.IPFlow.Key())
		channels = endpoint.NetworkServerEndpoint.TrafficChannels()
		return
	}

	// If there is no server endpoint listening, we first have to set one up.
	newEndpoint := m.specificManager.NewServer(flow.NetFlow, shila.TrafficNetworkEndpoint)
	if error = newEndpoint.SetupAndRun(); error != nil {
		return
	}

	// Add the endpoint to the mapping
	endpoint = shila.NewNetworkServerEndpointIPFlowRegister(newEndpoint)
	m.serverTrafficEndpoints[key] = endpoint

	// Register the new IP flow at the server endpoint
	endpoint.Register(flow.IPFlow.Key())

	// Announce the new traffic channels to the working side
	channels = endpoint.TrafficChannels()
	m.workingSide <- shila.PacketChannelAnnouncement{Announcer: newEndpoint, Channel: channels.Ingress}

	return
}

func (m *Manager) EstablishNewContactingClientEndpoint(flow shila.Flow) (contactingNetFlow shila.NetFlow, channels shila.PacketChannels, error error) {

	contactingNetFlow  	= shila.NetFlow{}
	channels 			= shila.PacketChannels{}
	error    			= nil

	if m.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}.")); return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clientContactingEndpoints[flow.IPFlow.Key()]; ok {
		error = shila.CriticalError(fmt.Sprint("Endpoint w/ key {", flow.IPFlow.Key(), "} already exists.")); return
	}

	// Establish a new contacting client endpoint
	contactingNetFlow 		  = m.specificManager.RemoteContactingFlow(flow.NetFlow)	// Does not contain src network address at this point.
	contactingEndpoint 		 := m.specificManager.NewClient(contactingNetFlow, shila.ContactingNetworkEndpoint)
	contactingNetFlow, error  = contactingEndpoint.SetupAndRun()						// Does now contain the src address as well.

	if error != nil {
		return
	}

	// Add it to the corresponding mapping
	m.clientContactingEndpoints[flow.IPFlow.Key()] = contactingEndpoint

	// Announce the new traffic channels to the working side
	channels = contactingEndpoint.TrafficChannels()
	m.workingSide <- shila.PacketChannelAnnouncement{Announcer: contactingEndpoint, Channel: channels.Ingress}

	return
}

func (m *Manager) EstablishNewTrafficClientEndpoint(flow shila.Flow) (trafficNetFlow shila.NetFlow, channels shila.PacketChannels, error error) {

	trafficNetFlow  	= shila.NetFlow{}
	channels 			= shila.PacketChannels{}
	error    			= nil

	if m.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}.")); return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.clientTrafficEndpoints[flow.IPFlow.Key()]; ok {
		error = shila.CriticalError(fmt.Sprint("Endpoint w/ key {", flow.IPFlow.Key(), "} already exists.")); return
	}

	trafficEndpoint := m.specificManager.NewClient(flow.NetFlow, shila.TrafficNetworkEndpoint)

	// Wait a certain amount of time to give the server endpoint time to establish itself
	time.Sleep(time.Duration(Config.WaitingTimeTrafficConnEstablishment) * time.Second)
	// TODO: Configuration parameter: Waiting time traffic conn establishment

	trafficNetFlow, error = trafficEndpoint.SetupAndRun()
	if error != nil {
		return
	}

	// Add it to the corresponding mapping
	m.clientTrafficEndpoints[flow.IPFlow.Key()] = trafficEndpoint

	// Announce the new traffic channels to the working side
	channels = trafficEndpoint.TrafficChannels()
	m.workingSide <- shila.PacketChannelAnnouncement{Announcer: trafficEndpoint, Channel: channels.Ingress}

	// Note:
	// The removal of the corresponding client contacting endpoint is triggered by the connTr
	// itself after obtaining the lock to change its state to established. Otherwise we are in danger
	// that we close the endpoint to early.

	return
}

func (m *Manager) TeardownTrafficSeverEndpoint(flow shila.Flow) error {

	if m.state.Not(shila.Running) {
		return  shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key     := shila.GetNetworkAddressKey(flow.NetFlow.Src)
	if ep, ok := m.serverTrafficEndpoints[key]; ok {
		ep.Unregister(flow.IPFlow.Key())
		if ep.IsEmpty() {
			err := ep.TearDown()
			delete(m.serverTrafficEndpoints, key)
			return err
		}
	}

	return nil
}

func (m *Manager) TeardownContactingClientEndpoint(ipFlow shila.IPFlow) error {

	if m.state.Not(shila.Running) {
		return  shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
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

	if m.state.Not(shila.Running) {
		return  shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
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