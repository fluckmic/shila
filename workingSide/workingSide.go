//
package workingSide

import (
	"fmt"
	"shila/core/connection"
	"shila/core/shila"
	"shila/shutdown"
)

type Manager struct {
	workType           WorkType
	connections        connection.Mapping
	trafficChannelPubs shila.PacketChannelPubChannel
	endpointIssues     shila.EndpointIssuePubChannel
}

func New(connections connection.Mapping, trafficChannelPubs shila.PacketChannelPubChannel,
	endpointIssues shila.EndpointIssuePubChannel, workType WorkType) *Manager {
	return &Manager{
		workType:           workType,
		connections:        connections,
		trafficChannelPubs: trafficChannelPubs,
		endpointIssues:     endpointIssues,
	}
}

func (manager *Manager) Setup() error {
	return nil
}

func (manager *Manager) Start() error {

	shutdown.Check()

	go manager.packetWorker()
	go manager.issueHandler()

	return nil
}

func (manager *Manager) CleanUp() { }

func (manager *Manager) Says(str string) string {
	return  fmt.Sprint(manager.Identifier(), ": ", str)
}

func (manager *Manager) Identifier() string {
	return fmt.Sprint(manager.WorkType(), " Working Side")
}

func (manager *Manager) WorkType() WorkType {
	return manager.workType
}

type WorkType uint8

const (
	_                = iota
	Ingress WorkType = iota
	Egress
)

func (wt WorkType) String() string {
	switch wt {
	case Ingress: 	return "Ingress"
	case Egress: 	return "Egress"
	}
	return "Unknown"
}