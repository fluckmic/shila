package workingSide

import (
	"github.com/bclicn/color"
	"shila/config"
	"shila/core/shila"
	"shila/log"
	"shila/shutdown"
)

func (manager *Manager) packetWorker() {
	for trafficChannelPub := range manager.trafficChannelPubs {
		go manager.servePacketChannel(trafficChannelPub.Channel, config.Config.WorkingSide.NumberOfWorkerPerChannel)
	}
}

func (manager *Manager) servePacketChannel(buffer shila.PacketChannel, numberOfWorker int) {
	for id := 0; id < numberOfWorker; id++ {
		go manager.handlePacketChannel(buffer)
	}
}

func (manager *Manager) handlePacketChannel(buffer shila.PacketChannel) {
	for p := range buffer {
		manager.processPacketChannel(p)
	}
}

func (manager *Manager) processPacketChannel(p *shila.Packet) {

	// Get the corresponding connection and processes the packet..
	con := manager.connections.Retrieve(p.Flow)
	err := con.ProcessPacket(p)

	// Any error leads inevitably to the closing of the connection.
	// All later packet that are processed by the same connection are silently dropped.
	// The closed connection is removed after a while; after its removal a packet is might
	// abel to use the connection without any error.

	if err, ok := err.(shila.TolerableError); ok {
		log.Error.Print(shila.PrependError(err, color.Yellow("Tolerable error in packet processing.")))
	}

	if err, ok := err.(shila.CriticalError); ok {
		shutdown.Fatal(shila.PrependError(err,"Critical error in packet processing."))
	}

}