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
	trafficChannelPubs shila.PacketChannelPubChannel
	endpointIssues 	   shila.EndpointIssuePubChannel
	state              shila.EntityState
}

type EndpointMapping map[shila.IPAddressKey] *kernelEndpoint.Device

func New(trafficChannelPubs shila.PacketChannelPubChannel, endpointIssues shila.EndpointIssuePubChannel) *Manager {
	return &Manager{
		endpoints:          make(EndpointMapping),
		trafficChannelPubs: trafficChannelPubs,
		endpointIssues: 	endpointIssues,
		state:              shila.NewEntityState(),
	}
}

func (m *Manager) Setup() error {

	if m.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	// Setup the namespaces
	if err := m.setupNamespaces(); err != nil {
		_ = m.removeNamespaces()
		return shila.PrependError(err, "Unable to setup namespace.")
	}

	// Setup additional routing
	if err := m.setupAdditionalRouting(); err != nil {
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return shila.PrependError(err, "Unable to setup routing.")
	}
	// Create the kernel endpoints
	if err := m.addKernelEndpoints(); err != nil {
		m.clearKernelEndpoints()
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return shila.PrependError(err, "Unable to add kernel endpoints.")
	}

	// Setup the kernel endpoints
	if err := m.setupKernelEndpoints(); err != nil {
		_ = m.tearDownKernelEndpoints()
		m.clearKernelEndpoints()
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return shila.PrependError(err, "Unable to setup kernel endpoints.")
	}

	m.state.Set(shila.Initialized)

	return nil
}

func (m *Manager) Start() error {

	if m.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	log.Verbose.Println("Starting kernel side...")

	if err := m.startKernelEndpoints(); err != nil {
		return err
	}

	// Announce all the traffic channels to the working side
	for _, kerep := range m.endpoints {
		m.trafficChannelPubs <- shila.PacketChannelPub{Publisher: kerep, Channel: kerep.TrafficChannels().Ingress}
	}

	m.state.Set(shila.Running)

	log.Verbose.Println("Kernel side started.")

	return nil
}

func (m *Manager) CleanUp() error {

	log.Info.Println("Mopping up the kernel side..")
	m.state.Set(shila.TornDown)

	err := m.tearDownKernelEndpoints()
	m.clearKernelEndpoints()
	err = m.clearAdditionalRouting()
	err = m.removeNamespaces()

	log.Info.Println("Kernel side mopped.")

	return err
}

func (m *Manager) GetTrafficChannels(key shila.IPAddressKey) (shila.PacketChannels, bool) {
	if endpoint, ok := m.endpoints[key]; !ok {
		return shila.PacketChannels{}, false
	} else {
		return endpoint.TrafficChannels(), true
	}
}

func (m *Manager) setupKernelEndpoints() error {
	for _, kerep := range m.endpoints {
		if err := kerep.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) startKernelEndpoints() error {
	var err error = nil
	for _, kerep := range m.endpoints {
		if err = kerep.Start(); err != nil {
			_ = m.tearDownKernelEndpoints()
			return err
		}
	}
	return err
}

func (m *Manager) tearDownKernelEndpoints() error {
	var err error = nil
	for _, kerep := range m.endpoints {
			_ = kerep.TearDown()
	}
	return err
}

func (m *Manager) clearKernelEndpoints() {
	for k := range m.endpoints {
		delete(m.endpoints, k)
	}
}

func (m *Manager) setupNamespaces() error {

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

func (m *Manager) removeNamespaces() error {

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

func (m *Manager) addKernelEndpoints() error {

	// Add the ingress kernel endpoint.
	key := shila.GetIPAddressKey(Config.IngressIP)
	kerep := kernelEndpoint.New(1, Config.IngressNamespace, Config.IngressIP, m.endpointIssues)
	m.endpoints[key] = &kerep

	// Add the egress kernel endpoint(s).
	numberOfEndpointsAdded := uint8(0)
	for {
		ip := getRandomIP()
		key := shila.GetIPAddressKey(ip)
		if _, ok := m.endpoints[key]; !ok {

			number := numberOfEndpointsAdded + 2
			kerep := kernelEndpoint.New(number, Config.EgressNamespace, ip, m.endpointIssues)

			m.endpoints[key] = &kerep
			numberOfEndpointsAdded++
		}
		if numberOfEndpointsAdded == Config.NumberOfEgressInterfaces {
			break
		}
	}

	return nil
}

func (m *Manager) setupAdditionalRouting() error {

	// Restrict the use of MPTCP to the virtual devices

	// If the ingress and egress interfaces are isolated in its own and fresh namespace,
	// then there is just the local interface which could also try to participate in MPTCP.
	// However, if this is not the case, then there could possibly multiple interfaces which
	// also want to participate. // TODO: https://github.com/fluckmic/shila/issues/16

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

func (m *Manager) clearAdditionalRouting() error {

	// Roll back the restriction of the use of MPTCP to the virtual devices.

	// If the ingress and egress interfaces are isolated in its own and fresh namespace,
	// then there is just the local interface which could also try to participate in MPTCP.
	// However, if this is not the case, then there could possibly multiple interfaces which
	// also want to participate. // TODO: https://github.com/fluckmic/shila/issues/16

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