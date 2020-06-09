//
package networkSide

import (
	"fmt"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
	"sync"
)

// This part of the network is independent of the backbone protocol chosen.

type Manager struct {
	specificManager           SpecificManager
	contactServer             shila.NetworkServerEndpoint
	serverTrafficEndpoints    shila.MappingNetworkServerEndpoint
	clientContactingEndpoints shila.MappingNetworkClientEndpoint
	clientTrafficEndpoints    shila.MappingNetworkClientEndpoint
	trafficChannelPubs        shila.PacketChannelPubChannels
	endpointIssues            shila.EndpointIssuePubChannels
	lock                      sync.Mutex
	state                     shila.EntityState
}

func New(trafficChannelPubs shila.PacketChannelPubChannels, endpointIssues shila.EndpointIssuePubChannels) *Manager {
	return &Manager{
		specificManager: 			NewSpecificManager(),
		trafficChannelPubs: 		trafficChannelPubs,
		endpointIssues: 			endpointIssues,
		serverTrafficEndpoints: 	make(shila.MappingNetworkServerEndpoint),
		clientContactingEndpoints: 	make(shila.MappingNetworkClientEndpoint),
		clientTrafficEndpoints:		make(shila.MappingNetworkClientEndpoint),
		lock:        	 			sync.Mutex{},
		state:       	 			shila.NewEntityState(),
	}
}

func (m *Manager) Setup() error {

	if m.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	contactLocalAddr := m.specificManager.ContactLocalAddr()
	m.contactServer   = m.specificManager.NewServer(contactLocalAddr, shila.ContactingNetworkEndpoint, m.endpointIssues.Ingress)

	m.state.Set(shila.Initialized)
	return nil
}

func (m *Manager) Start() error {

	if m.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	if err := m.contactServer.SetupAndRun(); err != nil {
		return shila.PrependError(err, "Unable to establish contacting server.")
	}

	// Announce the traffic channels to the working side
	m.trafficChannelPubs.Ingress <- shila.PacketChannelPub{
										Publisher: m.contactServer,
										Channel: m.contactServer.TrafficChannels().Ingress,
									}

	m.state.Set(shila.Running)
	return nil
}

func (m *Manager) CleanUp() error {

	m.state.Set(shila.TornDown)
	var err error = nil

	err = m.tearDownAndRemoveClientContactingEndpoints()
	err = m.tearDownAndRemoveClientTrafficEndpoints()
	err = m.tearDownAndRemoveServerTrafficEndpoints()

	err = m.contactServer.TearDown()
	m.contactServer = nil

	return err
}

func (m *Manager) EstablishNewTrafficServerEndpoint(lAddress shila.NetworkAddress, IPFlowKey shila.IPFlowKey) (channels shila.PacketChannels, error error) {

	channels 			= shila.PacketChannels{}
	error    			= nil

	if m.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}.")); return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	key := shila.GetNetworkAddressKey(lAddress)
	endpoint, ok := m.serverTrafficEndpoints[key]
	if ok {
		endpoint.Register(IPFlowKey)
		channels = endpoint.NetworkServerEndpoint.TrafficChannels()
		return
	}

	// If there is no server endpoint listening, we first have to set one up.
	newEndpoint := m.specificManager.NewServer(lAddress, shila.TrafficNetworkEndpoint, m.endpointIssues.Ingress)
	if error = newEndpoint.SetupAndRun(); error != nil {
		return
	}

	// Add the endpoint to the mapping
	endpoint = shila.NewNetworkServerEndpointIPFlowRegister(newEndpoint)
	m.serverTrafficEndpoints[key] = endpoint

	// Register the new IP flow at the server endpoint
	endpoint.Register(IPFlowKey)

	// Announce the new traffic channels to the working side
	channels = endpoint.TrafficChannels()
	m.trafficChannelPubs.Ingress <- shila.PacketChannelPub{
										Publisher: newEndpoint,
										Channel: channels.Ingress,
									}
	return
}

func (m *Manager) EstablishNewContactingClientEndpoint(flow shila.Flow) (contactingNetFlow shila.NetFlow, channels shila.PacketChannels, error error) {

	constructingFlow 	:= shila.Flow{IPFlow: flow.IPFlow, Kind: flow.Kind}
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
	// FIXME: ContactRemoteAddr could also specify the path taken to the contact server network endpoint.
	var ContactRemoteAddr *snet.UDPAddr
	_ = ContactRemoteAddr
	// contactRemoteAddr = m.specificManager.ContactRemoteAddr()
	constructingFlow.NetFlow  = m.specificManager.ContactRemoteAddr(flow.NetFlow) // Does not contain src network address at this point.

	// contactEndpoint := m.specificManager.NewClient(contactRemoteAddr, ipFlow, shila.ContactingNetworkEndpoint, m.endpointIssues.Egress)
	contactingEndpoint 		 := m.specificManager.NewClient(constructingFlow, shila.ContactingNetworkEndpoint, m.endpointIssues.Egress)
	contactingNetFlow, error  = contactingEndpoint.SetupAndRun()						// Does now contain the src address as well.

	if error != nil {
		return
	}

	// Add it to the corresponding mapping
	m.clientContactingEndpoints[flow.IPFlow.Key()] = contactingEndpoint

	// Announce the new traffic channels to the working side
	channels = contactingEndpoint.TrafficChannels()
	pub := shila.PacketChannelPub{Publisher: contactingEndpoint, Channel: channels.Ingress}
	m.trafficChannelPubs.Egress <- pub

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

	trafficEndpoint := m.specificManager.NewClient(flow, shila.TrafficNetworkEndpoint, m.endpointIssues.Egress)
	trafficNetFlow, error = trafficEndpoint.SetupAndRun()
	if error != nil {
		return
	}

	// Add it to the corresponding mapping
	m.clientTrafficEndpoints[flow.IPFlow.Key()] = trafficEndpoint

	// Announce the new traffic channels to the working side
	channels = trafficEndpoint.TrafficChannels()
	pub := shila.PacketChannelPub{Publisher: trafficEndpoint, Channel: channels.Ingress}
	m.trafficChannelPubs.Egress <- pub

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