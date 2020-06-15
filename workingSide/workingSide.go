//
package workingSide

import (
	"shila/core/connection"
	"shila/core/shila"
	"shila/shutdown"
)

type Manager struct {
	label 				Label
	connections     	connection.Mapping
	trafficChannelPubs 	shila.PacketChannelPubChannel
	endpointIssues 	   	shila.EndpointIssuePubChannel
}

func New(connections connection.Mapping, trafficChannelPubs shila.PacketChannelPubChannel,
	endpointIssues shila.EndpointIssuePubChannel, label Label) *Manager {
	return &Manager{
		label: 				label,
		connections:     	connections,
		trafficChannelPubs: trafficChannelPubs,
		endpointIssues:    	endpointIssues,
	}
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) Start() error {

	shutdown.Check()

	go m.packetWorker()
	go m.issueHandler()

	return nil
}

func (m *Manager) CleanUp() { }