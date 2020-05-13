package workingSide

import (
	"shila/config"
	"shila/core/connection"
	"shila/core/model"
	"shila/log"
)

type Manager struct {
	config 						config.Config
	connections 				*connection.Mapping
	trafficChannelAnnouncements chan model.TrafficChannels
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, connections *connection.Mapping, trafficChannelAnnouncements chan model.TrafficChannels) *Manager {
	return &Manager{config, connections, trafficChannelAnnouncements}
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) CleanUp() {
	
}

func (m *Manager) Start() error {

	go func() {
		for ch := range m.trafficChannelAnnouncements {
			go m.serveChannel(ch.Ingress, ch.Key, ch.Label, "ingress", m.config.WorkingSide.NumberOfWorkerPerChannel)
			go m.serveChannel(ch.Egress,  ch.Key, ch.Label, "egress",  m.config.WorkingSide.NumberOfWorkerPerChannel)
		}
	}()

	return nil
}

func (m *Manager) serveChannel(buffer model.PacketChannel, endpointKey model.EndpointKey,
	endpointLabel model.EndpointLabel, direction string, numberOfWorker int) {
	for id := 0; id < numberOfWorker; id++ {
		go m.handleChannel(buffer, endpointKey, endpointLabel, direction, id)
	}
}

func (m *Manager) handleChannel(buffer model.PacketChannel, endpointKey model.EndpointKey,
	endpointLabel model.EndpointLabel, direction string, handlerId int) {
	log.Verbose.Print("Started ", direction, " ", endpointLabel, " packet channel handler #", handlerId, " for ", endpointKey, ".")
	for p := range buffer {
		m.processChannel(p)
	}
}

func (m *Manager) processChannel(p *model.Packet) {

	// Get the connection
	var con *connection.Connection
	if key, err := p.IPHeaderKey(); err != nil {
		log.Error.Panicln(err.Error())
	} else {
		con = m.connections.Retrieve(key)
	}

	// Process the packet
	if err := con.ProcessPacket(p); err != nil {
		log.Error.Panicln(err.Error())
	}
}

