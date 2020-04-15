package kersi

import (
	"fmt"
	"net"
	"shila/config"
	"shila/helper"
	"shila/kersi/kerep"
	"shila/log"
)

type Manager struct {
	config    config.Config
	endpoints map[string]*kerep.Device
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func New(config config.Config) *Manager {
	return &Manager{config, nil}
}

func (m *Manager) Setup() error {

	if m.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", "Already setup."))
	}

	kCfg := m.config.KernelSide

	// Setup the mapping holding the kernel endpoints
	m.endpoints = make(map[string]*kerep.Device)

	// Create the kernel endpoints
	// Ingress
	if err := m.addKernelEndpoints(kCfg.NIngressKerEp, kCfg.IngressNamespace, kCfg.StartingIngressSubnet); err != nil {
		m.clearKernelEndpoints()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}
	// Egress
	if err := m.addKernelEndpoints(kCfg.NEgressKerEp, kCfg.EgressNamespace, kCfg.StartingEgressSubnet); err != nil {
		m.clearKernelEndpoints()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}

	// Setup the namespaces
	if err := m.setupNamespaces(); err != nil {
		m.clearKernelEndpoints()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}

	// Setup the kernel endpoints
	if err := m.setupKernelEndpoints(); err != nil {
		m.tearDownKernelEndpoints()
		m.clearKernelEndpoints()
		m.removeNamespaces()
		return Error(fmt.Sprint("Unable to setup kernel side",
			" - ", err.Error()))
	}

	return nil
}

func (m *Manager) CleanUp() {

	m.tearDownKernelEndpoints()
	m.clearKernelEndpoints()
	m.removeNamespaces()

}

func (m *Manager) Start() error {

	if !m.IsSetup() {
		return Error(fmt.Sprint("Cannot start kernel side",
			" - ", "Kernel side not yet setup."))
	}

	if m.IsRunning() {
		return Error(fmt.Sprint("Cannot start kernel side",
			" - ", "Kernel side already running."))

	}

	return nil
}

func (m *Manager) Stop() error {

	if !m.IsSetup() {
		return Error(fmt.Sprint("Cannot stop kernel side",
			" - ", "Device not yet setup."))
	}

	if !m.IsRunning() {
		return Error(fmt.Sprint("Cannot stop kernel side",
			" - ", "Device is not running."))

	}

	return nil
}

func (m *Manager) IsSetup() bool {
	return len(m.endpoints) > 0
}

func (m *Manager) IsRunning() bool {
	return false
}

// clearKernelEndpoints just empties the mapping
// but does not deallocate the endpoints beforehand!
func (m *Manager) clearKernelEndpoints() {
	for k := range m.endpoints {
		delete(m.endpoints, k)
	}
}

func (m *Manager) tearDownKernelEndpoints() {
	for _, kerep := range m.endpoints {
		if kerep.IsSetup() {
			_ = kerep.TearDown()
		}
	}
}

func (m *Manager) stopKernelEndpoints() {
	for _, kerep := range m.endpoints {
		if kerep.IsRunning() {
			_ = kerep.Stop()
		}
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

func (m *Manager) setupNamespaces() error {

	// Create ingress namespace
	if m.config.KernelSide.IngressNamespace != nil {
		if err := helper.AddNamespace(m.config.KernelSide.IngressNamespace.Name); err != nil {
			return Error(fmt.Sprint("Unable to setup ingress namespace ",
				m.config.KernelSide.IngressNamespace.Name, " - ", err.Error()))
		}
	}

	// Create egress namespace
	if m.config.KernelSide.EgressNamespace != nil {
		if err := helper.AddNamespace(m.config.KernelSide.EgressNamespace.Name); err != nil {
			if m.config.KernelSide.IngressNamespace != nil {
				// Remove ingress namespace if setting up the egress namespace fails.
				_ = helper.DeleteNamespace(m.config.KernelSide.IngressNamespace.Name)
			}
			return Error(fmt.Sprint("Unable to setup egress namespace ",
				m.config.KernelSide.EgressNamespace.Name, " - ", err.Error()))
		}
	}

	return nil
}

func (m *Manager) removeNamespaces() {
	// Remove ingress namespace
	if m.config.KernelSide.IngressNamespace != nil {
		_ = helper.DeleteNamespace(m.config.KernelSide.IngressNamespace.Name)
	}
	// Remove egress namespace
	if m.config.KernelSide.EgressNamespace != nil {
		_ = helper.DeleteNamespace(m.config.KernelSide.EgressNamespace.Name)
	}
}

// TODO: ingress and egress buffer not yet provided!
func (m *Manager) addKernelEndpoints(n uint, ns *helper.Namespace, sn net.IPNet) error {

	if startIP := sn.IP.To4(); startIP == nil {
		return Error(fmt.Sprint("Invalid starting IP: ", sn.IP))
	} else {
		for i := 0; i < int(n); i++ {

			// First create the identifier..
			newKerepId := kerep.NewIdentifier(uint(len(m.endpoints)+1), ns,
				net.IPNet{IP: net.IPv4(startIP[0], startIP[1], startIP[2], startIP[3]+byte(i)), Mask: sn.Mask})

			// ..then create the kernel endpoint..
			newKerep := kerep.New(newKerepId, m.config.KernelEndpoint, nil, nil)

			// ..and add it to the mapping.
			newKerepKey := newKerepId.Key()
			if _, ok := m.endpoints[newKerepKey]; !ok {
				m.endpoints[newKerepId.Key()] = newKerep
				log.Verbose.Println("Added kernel endpoint:", newKerepKey)
			} else {
				// Cannot have two endpoints w/ the same key.
				return Error(fmt.Sprint("Kernel endpoint already exists: ", newKerepKey))
			}
		}
	}
	return nil
}
