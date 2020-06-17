//
package networkSide

import (
	"fmt"
	"shila/core/shila"
	"sync"
)

type Manager struct {
	specificManager           SpecificManager
	contactServer             shila.NetworkServerEndpoint
	serverTrafficEndpoints    shila.MappingNetworkServerEndpoint
	clientContactingEndpoints shila.MappingNetworkClientEndpoint
	clientTrafficEndpoints    shila.MappingNetworkClientEndpoint
	trafficChannelPubs        shila.PacketChannelPubChannels
	endpointIssues            shila.EndpointIssuePubChannels
	serverEndpointIssues	  shila.EndpointIssuePubChannel
	lock                      sync.Mutex
	state                     shila.EntityState
}

func New(trafficChannelPubs shila.PacketChannelPubChannels, endpointIssues shila.EndpointIssuePubChannels) *Manager {
	return &Manager{
		specificManager: 			NewSpecificManager(),
		trafficChannelPubs: 		trafficChannelPubs,
		endpointIssues: 			endpointIssues,
		serverEndpointIssues:		make(shila.EndpointIssuePubChannel),
		serverTrafficEndpoints: 	make(shila.MappingNetworkServerEndpoint),
		clientContactingEndpoints: 	make(shila.MappingNetworkClientEndpoint),
		clientTrafficEndpoints:		make(shila.MappingNetworkClientEndpoint),
		lock:        	 			sync.Mutex{},
		state:       	 			shila.NewEntityState(),
	}
}

func (manager *Manager) Setup() error {

	if manager.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	contactLocalAddr := manager.specificManager.ContactLocalAddr()
	manager.contactServer   = manager.specificManager.NewServer(contactLocalAddr, shila.ContactNetworkEndpoint, manager.serverEndpointIssues)

	manager.state.Set(shila.Initialized)
	return nil
}

func (manager *Manager) Start() error {

	if manager.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	// Start the error worker.
	go manager.errorHandler()

	if err := manager.contactServer.SetupAndRun(); err != nil {
		return shila.PrependError(err, "Unable to establish contacting server.")
	}

	// Announce the traffic channels to the working side
	manager.trafficChannelPubs.Ingress <- shila.PacketChannelPub{
										Publisher: manager.contactServer,
										Channel:   manager.contactServer.TrafficChannels().Ingress,
									}

	manager.state.Set(shila.Running)
	return nil
}

func (manager *Manager) CleanUp() error {

	manager.state.Set(shila.TornDown)
	var err error = nil

	err = manager.tearDownAndRemoveClientContactingEndpoints()
	err = manager.tearDownAndRemoveClientTrafficEndpoints()
	err = manager.tearDownAndRemoveServerTrafficEndpoints()

	err = manager.contactServer.TearDown()
	manager.contactServer = nil

	// As soon as all server network endpoints are torn down
	// the channel is no longer needed. (Shuts down the issue worker.)
	close(manager.serverEndpointIssues)

	return err
}

func (manager *Manager) EstablishNewTrafficServerEndpoint(lAddress shila.NetworkAddress, flowKey shila.IPFlowKey) (channels shila.PacketChannels, error error) {

	channels 			= shila.PacketChannels{}
	error    			= nil

	if manager.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}.")); return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	endpointKey := shila.GetNetworkAddressKey(lAddress)
	endpointWrapper, ok := manager.serverTrafficEndpoints[endpointKey]
	if ok {
		endpointWrapper.Register(flowKey)
		channels = endpointWrapper.NetworkServerEndpoint.TrafficChannels()
		return
	}

	// If there is no server endpointWrapper listening, we first have to set one up.
	newEndpoint := manager.specificManager.NewServer(lAddress, shila.TrafficNetworkEndpoint, manager.serverEndpointIssues)
	if error = newEndpoint.SetupAndRun(); error != nil {
		return
	}

	// Add the endpointWrapper to the mapping
	endpointWrapper = shila.NewNetworkServerEndpointIPFlowRegister(newEndpoint)
	manager.serverTrafficEndpoints[endpointKey] = endpointWrapper

	// Register the new IP flow at the server endpointWrapper
	endpointWrapper.Register(flowKey)

	// Announce the new traffic channels to the working side
	channels = endpointWrapper.TrafficChannels()
	manager.trafficChannelPubs.Ingress <- shila.PacketChannelPub{
										Publisher: newEndpoint,
										Channel: channels.Ingress,
									}
	return
}

func (manager *Manager) EstablishNewContactingClientEndpoint(flow shila.Flow) (contactingNetFlow shila.NetFlow, channels shila.PacketChannels, error error) {

	contactingNetFlow  	= shila.NetFlow{}
	channels 			= shila.PacketChannels{}
	error    			= nil

	if manager.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state ", manager.state, ".")); return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	ipFlowKey := flow.IPFlow.Key()
	if _, ok := manager.clientContactingEndpoints[ipFlowKey]; ok {
		error = shila.CriticalError(fmt.Sprint("Endpoint w/ key ", ipFlowKey, " already exists.")); return
	}

	// Establish a new contacting client endpoint
	contactRemoteAddr := manager.specificManager.ContactRemoteAddr(flow.NetFlow.Dst)
	ipFlow 			  := flow.IPFlow

	// contactEndpoint := manager.specificManager.NewClient(contactRemoteAddr, ipFlow, shila.ContactNetworkEndpoint, manager.endpointIssues.Egress)
	contactingEndpoint 		 := manager.specificManager.NewContactClient(contactRemoteAddr, ipFlow, manager.endpointIssues.Egress)
	contactingNetFlow, error  = contactingEndpoint.SetupAndRun()						// Does now contain the src address as well.

	if error != nil {
		return
	}

	// Add it to the corresponding mapping
	manager.clientContactingEndpoints[ipFlowKey] = contactingEndpoint

	// Announce the new traffic channels to the working side
	channels = contactingEndpoint.TrafficChannels()
	pub := shila.PacketChannelPub{Publisher: contactingEndpoint, Channel: channels.Ingress}
	manager.trafficChannelPubs.Egress <- pub

	return
}

func (manager *Manager) EstablishNewTrafficClientEndpoint(flow shila.Flow) (trafficNetFlow shila.NetFlow, channels shila.PacketChannels, error error) {

	trafficNetFlow  	= shila.NetFlow{}
	channels 			= shila.PacketChannels{}
	error    			= nil

	if manager.state.Not(shila.Running) {
		error = shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}.")); return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	ipFlowKey := flow.IPFlow.Key()
	if _, ok := manager.clientTrafficEndpoints[ipFlowKey]; ok {
		error = shila.CriticalError(fmt.Sprint("Endpoint w/ key ", ipFlowKey, " already exists.")); return
	}

	lAddrContact := flow.NetFlow.Src
	rAddr 		 := flow.NetFlow.Dst
	ipFlow 		 := flow.IPFlow

	trafficEndpoint := manager.specificManager.NewTrafficClient(lAddrContact, rAddr, ipFlow, manager.endpointIssues.Egress)

	trafficNetFlow, error = trafficEndpoint.SetupAndRun()
	if error != nil {
		return
	}

	// Add it to the corresponding mapping
	manager.clientTrafficEndpoints[ipFlowKey] = trafficEndpoint

	// Announce the new traffic channels to the working side
	channels = trafficEndpoint.TrafficChannels()
	pub := shila.PacketChannelPub{Publisher: trafficEndpoint, Channel: channels.Ingress}
	manager.trafficChannelPubs.Egress <- pub

	// Note:
	// The removal of the corresponding client contacting endpoint is triggered by the connTr
	// itself after obtaining the lock to change its state to established. Otherwise we are in danger
	// that we close the endpoint to early.

	return
}

func (manager *Manager) TeardownTrafficSeverEndpoint(flow shila.Flow) error {

	if manager.state.Not(shila.Running) {
		return  shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	key     := shila.GetNetworkAddressKey(flow.NetFlow.Src)
	if ep, ok := manager.serverTrafficEndpoints[key]; ok {
		ep.Unregister(flow.IPFlow.Key())
		if ep.IsEmpty() {
			err := ep.TearDown()
			delete(manager.serverTrafficEndpoints, key)
			return err
		}
	}

	return nil
}

func (manager *Manager) TeardownContactingClientEndpoint(ipFlow shila.IPFlow) error {

	if manager.state.Not(shila.Running) {
		return  shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	if ep, ok := manager.clientContactingEndpoints[ipFlow.Key()]; ok {
		err := ep.TearDown()
		delete(manager.clientContactingEndpoints, ipFlow.Key())
		return err
	}

	return nil
}

func (manager *Manager) TeardownTrafficClientEndpoint(ipFlow shila.IPFlow) error {

	if manager.state.Not(shila.Running) {
		return  shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	if ep, ok := manager.clientTrafficEndpoints[ipFlow.Key()]; ok {
		err := ep.TearDown()
		delete(manager.clientTrafficEndpoints, ipFlow.Key())
		return err
	}
	return nil
}

func (manager *Manager) tearDownAndRemoveServerTrafficEndpoints() error {
	var err error = nil
	for key, ep := range manager.serverTrafficEndpoints {
		err = ep.TearDown()
		delete(manager.serverTrafficEndpoints, key)
	}
	return err
}

func (manager *Manager) tearDownAndRemoveClientContactingEndpoints() error {
	var err error = nil
	for key, ep := range manager.clientContactingEndpoints {
		err = ep.TearDown()
		delete(manager.clientContactingEndpoints, key)
	}
	return err
}

func (manager *Manager) tearDownAndRemoveClientTrafficEndpoints() error {
	var err error = nil
	for key, ep := range manager.clientTrafficEndpoints {
		err = ep.TearDown()
		delete(manager.clientTrafficEndpoints, key)
	}
	return err
}