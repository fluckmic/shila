package kernelSide

import (
	"fmt"
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/kernelSide/ipCommand"
	"shila/kernelSide/kernelEndpoint"
	"shila/log"
)

type Manager struct {
	config    	config.Config
	endpoints 	EndpointMapping
	workingSide chan shila.PacketChannelAnnouncement
	state       shila.EntityState
}

type EndpointMapping map[shila.IPAddressKey] *kernelEndpoint.Device

func New(config config.Config, workingSide chan shila.PacketChannelAnnouncement) *Manager {
	return &Manager{
		config: 		config,
		endpoints: 		make(EndpointMapping),
		workingSide: 	workingSide,
		state: 			shila.NewEntityState(),
	}
}

func (m *Manager) Setup() error {

	if m.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	kCfg := m.config.KernelSide

	// Setup the namespaces
	if err := m.setupNamespaces(); err != nil {
		_ = m.removeNamespaces()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}

	// Setup additional routing
	if err := m.setupAdditionalRouting(); err != nil {
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}
	// Create the kernel endpoints
	// Egress
	if err := m.addKernelEndpoints(kCfg.NEgressKerEp, kCfg.EgressNamespace, kCfg.EgressIP); err != nil {
		m.clearKernelEndpoints()
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}
	// Ingress
	if err := m.addKernelEndpoints(1, kCfg.IngressNamespace, kCfg.IngressIP); err != nil {
		m.clearKernelEndpoints()
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}

	// Setup the kernel endpoints
	if err := m.setupKernelEndpoints(); err != nil {
		_ = m.tearDownKernelEndpoints()
		m.clearKernelEndpoints()
		_ = m.clearAdditionalRouting()
		_ = m.removeNamespaces()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}

	m.state.Set(shila.Initialized)

	return nil
}

func (m *Manager) CleanUp() error {

	var err error = nil

	log.Info.Println("Mopping up the kernel side..")

	err = m.tearDownKernelEndpoints()
	m.clearKernelEndpoints()
	err = m.clearAdditionalRouting()
	err = m.removeNamespaces()

	m.state.Set(shila.TornDown)

	log.Info.Println("Kernel side mopped.")

	return err
}

func (m *Manager) Start() error {

	if m.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", m.state, "}."))
	}

	log.Verbose.Println("Starting kernel side...")

	// Announce all the traffic channels to the working side
	for _, kerep := range m.endpoints {
		m.workingSide <- shila.PacketChannelAnnouncement{Announcer: kerep, Channel: kerep.TrafficChannels().Ingress}
	}

	m.state.Set(shila.Running)

	log.Verbose.Println("Kernel side started.")

	return nil
}

func (m *Manager) setupKernelEndpoints() error {
	for _, kerep := range m.endpoints {
		if err := kerep.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetTrafficChannels(key shila.IPAddressKey) (shila.PacketChannels, bool) {
	if endpoint, ok := m.endpoints[key]; !ok {
		return shila.PacketChannels{}, false
	} else {
		return endpoint.TrafficChannels(), true
	}
}

// clearKernelEndpoints just empties the mapping
// but does not deallocate the endpoints beforehand!
func (m *Manager) clearKernelEndpoints() {
	for k := range m.endpoints {
		delete(m.endpoints, k)
	}
}

func (m *Manager) tearDownKernelEndpoints() error {
	var err error = nil
	for _, kerep := range m.endpoints {
		if kerep.IsSetup() {
			_ = kerep.TearDown()
		}
	}
	return err
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

func (m *Manager) setupNamespaces() error {

	// Create ingress namespace
	if m.config.KernelSide.IngressNamespace != nil {
		if err := ipCommand.AddNamespace(m.config.KernelSide.IngressNamespace.Name); err != nil {
			return Error(fmt.Sprint("Unable to setup ingress namespace ",
				m.config.KernelSide.IngressNamespace.Name, " - ", err.Error()))
		}
	}

	// Create egress namespace
	if m.config.KernelSide.EgressNamespace != nil {
		if err := ipCommand.AddNamespace(m.config.KernelSide.EgressNamespace.Name); err != nil {
			return Error(fmt.Sprint("Unable to setup egress namespace ",
				m.config.KernelSide.EgressNamespace.Name, " - ", err.Error()))
		}
	}

	return nil
}

func (m *Manager) removeNamespaces() error {

	var err error = nil

	// Remove ingress namespace
	if m.config.KernelSide.IngressNamespace != nil {
		err = ipCommand.DeleteNamespace(m.config.KernelSide.IngressNamespace.Name)
	}
	// Remove egress namespace
	if m.config.KernelSide.EgressNamespace != nil {
		err = ipCommand.DeleteNamespace(m.config.KernelSide.EgressNamespace.Name)
	}

	return err
}

func (m *Manager) addKernelEndpoints(n uint, ns *ipCommand.Namespace, ip net.IP) error {

	if startIP := ip.To4(); startIP == nil {
		return Error(fmt.Sprint("Invalid starting IP: ", ip))
	} else {
		for i := 0; i < int(n); i++ {

			// First create the identifier..
			newKerepId := kernelEndpoint.NewIdentifier(uint(len(m.endpoints)+1), ns,
				net.IPv4(startIP[0], startIP[1], startIP[2], startIP[3]+byte(i)))

			// ..then create the kernel endpoint..
			newKerep := kernelEndpoint.New(newKerepId, m.config.KernelEndpoint)

			// ..and add it to the mapping.
			newKerepKey := newKerepId.Key()
			if _, ok := m.endpoints[newKerepKey]; !ok {
				m.endpoints[newKerepId.Key()] = newKerep
				// log.Verbose.Print("Added kernel endpoint: ", newKerepKey, ".")
			} else {
				// Cannot have two endpoints w/ the same key.
				return Error(fmt.Sprint("Kernel endpoint already exists: ", newKerepKey))
			}
		}
	}
	return nil
}

func (m *Manager) setupAdditionalRouting() error {

	kCfg := m.config.KernelSide

	// Restrict the use of MPTCP to the virtual devices
	// If the ingress and egress interfaces are isolated in its own and fresh namespace,
	// then there is just the local interface which could also try to participate in MPTCP.
	// However, if this is not the case, then there could possibly multiple interfaces which
	// also want to participate. // TODO: handle these cases.
	// ip link set dev lo multipath off
	args := []string{"link", "set", "dev", "lo", "multipath", "off"}
	if err := ipCommand.Execute(kCfg.IngressNamespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup additional routing.", " - ", err.Error()))
	}
	if err := ipCommand.Execute(kCfg.EgressNamespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup additional routing.", " - ", err.Error()))
	}

	// SYN packets coming from client side connect calls are sent from the
	// local interface, route them through one of the egress devices
	// ip rule add to <ip> iif lo table <id>
	// TODO: I dont like the disconnection between table 1 here and the one used later..
	args = []string{"rule", "add", "to", kCfg.IngressIP.String(), "table", "1"}
	if err := ipCommand.Execute(kCfg.EgressNamespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup additional routing.", " - ", err.Error()))
	}

	return nil
}

func (m *Manager) clearAdditionalRouting() error {

	kCfg := m.config.KernelSide

	// Roll back the restriction of the use of MPTCP to the virtual devices.
	// If the ingress and egress interfaces are isolated in its own and fresh namespace,
	// then there is just the local interface which could also try to participate in MPTCP.
	// However, if this is not the case, then there could possibly multiple interfaces which
	// also want to participate. // TODO: handle these cases.
	// ip link set dev lo multipath on
	args := []string{"link", "set", "dev", "lo", "multipath", "on"}
	err := ipCommand.Execute(kCfg.IngressNamespace, args...)
	err = ipCommand.Execute(kCfg.EgressNamespace, args...)

	// SYN packets coming from client side connect calls are sent from the
	// local interface, route them through one of the egress devices.
	// ip rule add to <ip> iif lo table <id>
	// TODO: I dont like the disconnection between table 1 here and the one used later..
	args = []string{"rule", "delete", "to", kCfg.IngressIP.String(), "table", "1"}
	err = ipCommand.Execute(kCfg.EgressNamespace, args...)

	return err
}
