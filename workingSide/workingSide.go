package workingSide

import (
	"shila/config"
	"shila/kernelSide"
	"shila/log"
	"shila/networkSide"
	"shila/shila"
	"shila/shila/connection"
)

type Manager struct {
	config 		config.Config
	kernelSide  *kernelSide.Manager
	networkSide *networkSide.Manager
	connections *connection.Mapping
}

type Error string

func New(config config.Config, kernelSide *kernelSide.Manager,
	networkSide *networkSide.Manager,	connections *connection.Mapping) *Manager {
	return &Manager{config, kernelSide, networkSide, connections}
}

func (e Error) Error() string {
	return string(e)
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) CleanUp() {
	
}

func (m *Manager) Start() error {

	// For the kernel side, the number of kernel endpoints is fixed.
	// So we can start all handler right from the beginning.
	for kernelEndpointKey, kernelEndpoint := range *m.kernelSide.GetEndpoints() {
		go m.serveChannel(kernelEndpoint.TrafficChannels().Ingress,
			m.config.WorkingSide.NumberOfKernelEndpointIngressHandler, string(kernelEndpointKey))
		go m.serveChannel(kernelEndpoint.TrafficChannels().Egress,
			m.config.WorkingSide.NumberOfKernelEndpointEgressHandler, string(kernelEndpointKey))
	}

	/*
	// For the network side, in the beginning there is just the contacting channel
	contactingSererNetworkEndpoint := m.networkSide.GetContactingServerEndpoint()
	go m.serveChannel(contactingSererNetworkEndpoint.TrafficChannels().Ingress,
		m.config.WorkingSide.NumberOfNetworkEndpointIngressHandler, "contacting server")
	go m.serveChannel(contactingSererNetworkEndpoint.TrafficChannels().Egress,
		m.config.WorkingSide.NumberOfNetworkEndpointEgressHandler, "contacting server")
	*/

	// Start listener to be ready to start and terminate worker for new established connections

	return nil
}

func (m *Manager) serveChannel(buffer shila.PacketChannel, numberOfWorker int,  endpointKey string) {
	for id := 0; id < numberOfWorker; id++ {
		go m.handleChannel(buffer, endpointKey, id)
	}
}

func (m *Manager) handleChannel(buffer shila.PacketChannel, endpointKey string, handlerId int) {
	log.Verbose.Print("Started packet channel handler ", endpointKey, "-", handlerId, ".")
	for p := range buffer {
		m.processChannel(p)
	}
}

func (m *Manager) processChannel(p *shila.Packet) {

	// Get the connection
	var con *connection.Connection
	if key, err := p.Key(); err != nil {
		log.Error.Panicln(err.Error())
	} else {
		con = m.connections.Retrieve(connection.ID(key))
	}

	// Process the packet
	if err := con.ProcessPacket(p); err != nil {
		log.Error.Panicln(err.Error())
	}
}

