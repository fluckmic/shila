package workingSide

import (
	"shila/config"
	"shila/core/connection"
	"shila/core/shila"
	"shila/log"
)

type Manager struct {
	config 						config.Config
	connections 				*connection.Mapping
	trafficChannelAnnouncements chan shila.PacketChannelAnnouncement
}

type Error string
func (e Error) Error() string {
	return string(e)
}

func New(config config.Config, connections *connection.Mapping, trafficChannelAnnouncements chan shila.PacketChannelAnnouncement) *Manager {
	return &Manager{config, connections, trafficChannelAnnouncements}
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) CleanUp() {
	
}

func (m *Manager) Start() error {

	go func() {
		for anc := range m.trafficChannelAnnouncements {
			// log.Verbose.Print("Working side received announcement for new traffic channel {", anc.Announcer.Key()," ",anc.Announcer.Label(),"}.")
			go m.serveChannel(anc.Channel, m.config.WorkingSide.NumberOfWorkerPerChannel)
		}
	}()

	return nil
}

func (m *Manager) serveChannel(buffer shila.PacketChannel, numberOfWorker int) {
	for id := 0; id < numberOfWorker; id++ {
		go m.handleChannel(buffer)
	}
}

func (m *Manager) handleChannel(buffer shila.PacketChannel) {
	for p := range buffer {
		m.processChannel(p)
	}
}

func (m *Manager) processChannel(p *shila.Packet) {
	// Get the connection and process the packet
	con := m.connections.Retrieve(p.Flow)
	if err := con.ProcessPacket(p); err != nil {
		log.Error.Panicln(err.Error())
	}
}

