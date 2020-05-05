package networkSide

import (
	"fmt"
	"shila/config"
	"shila/log"
	"shila/networkSide/networkEndpoint"
	"shila/shila"
)

type Manager struct {
	config    			config.Config
	contactingServer	shila.ServerNetworkEndpoint
	serverEndpoints		ServerEndpointMapping
	clientEndpoints		ClientEndpointMapping
	isRunning			bool
}

type _NetworkAddress_ string
type ServerEndpointMapping map[_NetworkAddress_] shila.ServerNetworkEndpoint
type _NetworkAddressAndPath_ string
type ClientEndpointMapping map[_NetworkAddressAndPath_] shila.ClientNetworkEndpoint

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config) *Manager {
	return &Manager{config,nil,nil,nil, false}
}

func (m *Manager) Setup() error {

	if m.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel side - Already setup."))
	}

	// Create the contacting server
	addr := networkEndpoint.Generator{}.NewLocalAddress(m.config.NetworkSide.ContactingServerPort)
	m.contactingServer = networkEndpoint.Generator{}.NewServer(addr, shila.ContactingNetworkEndpoint, m.config.NetworkEndpoint)

	// Create the mappings
	m.serverEndpoints = make(ServerEndpointMapping)
	m.clientEndpoints = make(ClientEndpointMapping)

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

	m.isRunning = true

	log.Verbose.Println("Network side started.")

	return nil
}

func (m *Manager) CleanUp() {

}

func (m *Manager) EstablishNewServerEndpoint(addr shila.NetworkAddress) (shila.TrafficChannels, error) {

	// If there already exists a server endpoint listening for addr, return its channels
	if ep, ok := m.serverEndpoints[_NetworkAddress_(addr.String())]; ok {
		return ep.TrafficChannels(), nil
	} else {
	// Otherwise establish a new one
		newServerEndpoint := networkEndpoint.Generator{}.NewServer(addr, shila.TrafficNetworkEndpoint, m.config.NetworkEndpoint)
		if err := newServerEndpoint.SetupAndRun(); err != nil {
			return shila.TrafficChannels{}, Error(fmt.Sprint("Unable to establish new server endpoint - ", err.Error()))
		}

		// TODO: Inform worker about new channels!

		return newServerEndpoint.TrafficChannels(), nil
	}
}

func (m *Manager) EstablishNewContactingClientEndpoint(addr shila.NetworkAddress) (shila.TrafficChannels, error) {

	/*
		// TODO: Config!
		egress  :=  make(shila.PacketChannel, 10)
		ingress :=  make(shila.PacketChannel, 10)
		// TODO: "default" path here, since this is the connection used for the contacting
		c.channels.Contacting.Endpoint = networkEndpoint.Generator{}.NewClient(c.header.Dst, c.header.Path, shila.ContactingNetworkEndpoint,
			shila.TrafficChannels{Ingress: c.channels.Contacting.Channels.Egress, Egress: c.channels.Contacting.Channels.Ingress})

		if err := c.channels.Contacting.Endpoint.SetupAndRun(); err != nil {
			return Error(fmt.Sprint("Failed to establish contacting channel - ", err.Error()))
		}
	*/
	return shila.TrafficChannels{}, nil
}

func (m *Manager) EstablishNewTrafficClientEndpoint(addr shila.NetworkAddress, path shila.NetworkPath) (shila.TrafficChannels, error) {

	// Close the contacting client endpoint as soon as the traffic client endpoint is established.

	return shila.TrafficChannels{}, nil
}

func (m *Manager) GetContactingServerEndpoint() shila.ServerNetworkEndpoint {
	return m.contactingServer
}

func (m *Manager) IsSetup() bool {
	return m.contactingServer != nil
}

func (m *Manager) IsRunning() bool {
	return m.isRunning
}
