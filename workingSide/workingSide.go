package workingSide

import (
	"shila/core/connection"
	"shila/core/shila"
	"shila/log"
)

type Manager struct {
	connections     connection.Mapping
	trafficChannels chan shila.PacketChannelAnnouncement
}

func New(connections connection.Mapping, trafficChannels chan shila.PacketChannelAnnouncement) *Manager {
	return &Manager{
		connections:     connections,
		trafficChannels: trafficChannels,
	}
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) Start() error {
	go func() {
		for anc := range m.trafficChannels {
			log.Verbose.Print("Working side received announcement for new traffic channel {", anc.Announcer.Key(), ",", anc.Announcer.Label(), "}.")
			go m.serveChannel(anc.Channel, Config.NumberOfWorkerPerChannel)
		}
	}()
	return nil
}

func (m *Manager) CleanUp() { }

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

	// Get the corresponding connection and processes the packet..
	con := m.connections.Retrieve(p.Flow)

	err := con.ProcessPacket(p)

	// Any error leads inevitably to the closing of the connection.
	// All later packet that are processed by the same connection are silently dropped.
	// The closed connection is removed after a while; after its removal a packet is might
	// abel to use the packet without any error.

	switch err := err.(type) {
	case shila.ThirdPartyError: 	log.Info.Print(err.Error())		// Really not our fault.
	case shila.TolerableError:  	log.Info.Panic(err.Error())		// Probably our fault.
	case shila.CriticalError:		log.Error.Panic(err.Error()) 	// Most likely our fault.
	default:						return
	}
}

