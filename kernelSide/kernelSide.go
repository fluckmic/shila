//
package kernelSide

import (
	"fmt"
	"math/rand"
	"net"
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint"
	"shila/kernelSide/network"
	"shila/log"
	"time"
)

// The table number depends on the number assigned to the virtual interface.
// The ingress interface has number 1, that is, the first egress interface has number 2.
const tableNumberOfFirstEgressInterface = "2"

type Manager struct {
	endpoints          EndpointMapping
	trafficChannelPubs shila.PacketChannelPubChannels
	endpointIssues 	   shila.EndpointIssuePubChannel
	state              shila.EntityState
}

type EndpointMapping map[shila.IPAddressKey] *kernelEndpoint.Device

func New(trafficChannelPubs shila.PacketChannelPubChannels) *Manager {
	return &Manager{
		trafficChannelPubs: trafficChannelPubs,
		endpoints:          make(EndpointMapping),
		endpointIssues: 	make(shila.EndpointIssuePubChannel),
		state:              shila.NewEntityState(),
	}
}

func (manager *Manager) Setup() error {

	if manager.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	// Setup the namespaces
	if err := manager.setupNamespaces(); err != nil {
		_ = manager.removeNamespaces()
		return shila.PrependError(err, "Unable to setup namespace.")
	}

	// Setup additional routing
	if err := manager.setupAdditionalRouting(); err != nil {
		_ = manager.clearAdditionalRouting()
		_ = manager.removeNamespaces()
		return shila.PrependError(err, "Unable to setup routing.")
	}
	// Create the kernel endpoints
	if err := manager.addKernelEndpoints(); err != nil {
		manager.clearKernelEndpoints()
		_ = manager.clearAdditionalRouting()
		_ = manager.removeNamespaces()
		return shila.PrependError(err, "Unable to add kernel endpoints.")
	}

	// Setup the kernel endpoints
	if err := manager.setupKernelEndpoints(); err != nil {
		_ = manager.tearDownKernelEndpoints()
		manager.clearKernelEndpoints()
		_ = manager.clearAdditionalRouting()
		_ = manager.removeNamespaces()
		return shila.PrependError(err, "Unable to setup kernel endpoints.")
	}

	manager.state.Set(shila.Initialized)

	return nil
}

func (manager *Manager) Start() error {

	if manager.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", manager.state, "}."))
	}

	// Start the error handler
	go manager.errorHandler()

	if err := manager.startKernelEndpoints(); err != nil {
		return err
	}

	// Announce all the traffic channels to the working side
	for _, kerep := range manager.endpoints {
		pub := shila.PacketChannelPub{Publisher: kerep, Channel: kerep.TrafficChannels().Ingress}
		if 		  kerep.Role() == shila.EgressKernelEndpoint {
			manager.trafficChannelPubs.Egress <- pub
		} else if kerep.Role() == shila.IngressKernelEndpoint {
			manager.trafficChannelPubs.Ingress <- pub
		} else {
			return shila.CriticalError(fmt.Sprint("Invalid kernel endpoint label {", kerep.Role(), "},"))
		}

	}

	manager.state.Set(shila.Running)

	log.Verbose.Println("Kernel side started.")

	return nil
}

func (manager *Manager) CleanUp() error {

	manager.state.Set(shila.TornDown)

	err := manager.tearDownKernelEndpoints()
	manager.clearKernelEndpoints()
	err = manager.clearAdditionalRouting()
	err = manager.removeNamespaces()

	close(manager.endpointIssues)

	return err
}

func (manager *Manager) GetTrafficChannels(key shila.IPAddressKey) (shila.PacketChannels, bool) {
	if endpoint, ok := manager.endpoints[key]; !ok {
		return shila.PacketChannels{}, false
	} else {
		return endpoint.TrafficChannels(), true
	}
}

func (manager *Manager) setupKernelEndpoints() error {
	for _, kerep := range manager.endpoints {
		if err := kerep.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) startKernelEndpoints() error {
	var err error = nil
	for _, kerep := range manager.endpoints {
		if err = kerep.Start(); err != nil {
			_ = manager.tearDownKernelEndpoints()
			return err
		}
	}
	return err
}

func (manager *Manager) tearDownKernelEndpoints() error {
	var err error = nil
	for _, kerep := range manager.endpoints {
			_ = kerep.TearDown()
	}
	return err
}

func (manager *Manager) clearKernelEndpoints() {
	for k := range manager.endpoints {
		delete(manager.endpoints, k)
	}
}

func (manager *Manager) setupNamespaces() error {

	// Create ingress namespace
	if Config.IngressNamespace.NonEmpty {
		if err := network.AddNamespace(Config.IngressNamespace); err != nil {
			return err
		}
	}

	// Create egress namespace
	if Config.EgressNamespace.NonEmpty {
		if err := network.AddNamespace(Config.EgressNamespace); err != nil {
			return err
		}
	}

	return nil
}

func (manager *Manager) removeNamespaces() error {

	var err error = nil

	// Remove ingress namespace
	if Config.IngressNamespace.NonEmpty {
		err = network.DeleteNamespace(Config.IngressNamespace)
	}
	// Remove egress namespace
	if Config.EgressNamespace.NonEmpty {
		err = network.DeleteNamespace(Config.EgressNamespace)
	}

	return err
}

func (manager *Manager) addKernelEndpoints() error {

	// Add the ingress kernel endpoint.
	key := shila.GetIPAddressKey(Config.IngressIP)
	kerep := kernelEndpoint.New(1, Config.IngressNamespace, Config.IngressIP, shila.IngressKernelEndpoint, manager.endpointIssues)
	manager.endpoints[key] = &kerep

	// Add the egress kernel endpoint(s).
	numberOfEndpointsAdded := uint8(0)
	for {
		ip := getRandomIP()
		key := shila.GetIPAddressKey(ip)
		if _, ok := manager.endpoints[key]; !ok {

			number := numberOfEndpointsAdded + 2
			kerep := kernelEndpoint.New(number, Config.EgressNamespace, ip, shila.EgressKernelEndpoint, manager.endpointIssues)

			manager.endpoints[key] = &kerep
			numberOfEndpointsAdded++
		}
		if numberOfEndpointsAdded == Config.NumberOfEgressInterfaces {
			break
		}
	}

	return nil
}

func (manager *Manager) setupAdditionalRouting() error {

	// Restrict the use of MPTCP to the virtual devices

	// If the ingress and egress interfaces are isolated in its own and fresh namespace,
	// then there is just the local interface which could also try to participate in MPTCP.
	// However, if this is not the case, then there could possibly multiple interfaces which
	// also want to participate. // TODO.

	// ip link set dev lo multipath off
	args := []string{"link", "set", "dev", "lo", "multipath", "off"}
	if err := network.Execute(Config.IngressNamespace, args...); err != nil {
		return err
	}
	if err := network.Execute(Config.EgressNamespace, args...); err != nil {
		return err
	}

	// SYN packets coming from client side connect calls are sent from the
	// local interface, route them through one of the egress devices..

	// ip rule add to <ip> iif lo table <id>
	args = []string{"rule", "add", "to", Config.IngressIP.String(), "table", tableNumberOfFirstEgressInterface}
	if err := network.Execute(Config.EgressNamespace, args...); err != nil {
		return err
	}

	return nil
}

func (manager *Manager) clearAdditionalRouting() error {

	// Roll back the restriction of the use of MPTCP to the virtual devices.

	// If the ingress and egress interfaces are isolated in its own and fresh namespace,
	// then there is just the local interface which could also try to participate in MPTCP.
	// However, if this is not the case, then there could possibly multiple interfaces which
	// also want to participate. // TODO.

	// ip link set dev lo multipath on
	args := []string{"link", "set", "dev", "lo", "multipath", "on"}
	err := network.Execute(Config.IngressNamespace, args...)
	err = network.Execute(Config.EgressNamespace, args...)

	// ip rule add to <ip> iif lo table <id>
	args = []string{"rule", "delete", "to", Config.IngressIP.String(), "table", tableNumberOfFirstEgressInterface}
	err = network.Execute(Config.EgressNamespace, args...)

	return err
}

func getRandomIP() net.IP {
		rand.Seed(time.Now().Unix())
		return net.IPv4(
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)))
}