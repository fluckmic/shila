package workingSide

import (
	"shila/core/shila"
	"shila/shutdown"
)

func (m *Manager) packetWorker() {
	for trafficChannelPub := range m.trafficChannelPubs {
		go m.servePacketChannel(trafficChannelPub.Channel, Config.NumberOfWorkerPerChannel)
	}
}

func (m *Manager) servePacketChannel(buffer shila.PacketChannel, numberOfWorker int) {
	for id := 0; id < numberOfWorker; id++ {
		go m.handlePacketChannel(buffer)
	}
}

func (m *Manager) handlePacketChannel(buffer shila.PacketChannel) {
	for p := range buffer {
		m.processPacketChannel(p)
	}
}

func (m *Manager) processPacketChannel(p *shila.Packet) {

	// Get the corresponding connection and processes the packet..
	con := m.connections.Retrieve(p.Flow)
	err := con.ProcessPacket(p)

	// Any error leads inevitably to the closing of the connection.
	// All later packet that are processed by the same connection are silently dropped.
	// The closed connection is removed after a while; after its removal a packet is might
	// abel to use the packet without any error.

	if err, ok := err.(shila.CriticalError); ok {
		shutdown.Fatal(shila.PrependError(err, "Critical error in packet processing."))
	}

}