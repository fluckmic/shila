package stumps

import (
	"shila/kersi"
	"shila/log"
	"shila/parser"
	"shila/shila"
)

// Will be part of the core config
var nKerepIngressHandler = 2

var kernelSide *kersi.Manager

func Setup(kersi *kersi.Manager) error {
	kernelSide = kersi
	return nil
}

func Start() error {
	for kerepKey, ep := range kernelSide.Endpoints {
		go serveKerepIngress(ep.Channels.Ingress, kerepKey)
	}
	return nil
}

func serveKerepIngress(buffer chan *shila.Packet, kerepKey string) {
	for id := 0; id < nKerepIngressHandler; id++ {
		go handleKerepIngress(buffer, kerepKey, id)
	}
}

func handleKerepIngress(buffer chan *shila.Packet, kerepKey string, handlerId int) {
	log.Verbose.Print("Started kernel endpoint ingress listener ", kerepKey, "-", handlerId, ".")
	for p := range buffer {
		processKerepIngress(p)
	}
}

func processKerepIngress(p *shila.Packet) {

	// Get the flow of the packet.

	// Decode the IPv4 and the TCP layer of the packet
	if err := parser.DecodeIPv4andTCPLayer(p); err != nil {
		log.Error.Panicln(err.Error())
	}

	// Decode the IPv4 options of the packet
	if err := parser.DecodeIPv4Options(p); err != nil {
		log.Error.Panicln(err.Error())
	}

	// Decode the MPTCP options of the packet
	if err := parser.DecodeMPTCPOptions(p); err != nil {
		log.Error.Panicln(err.Error())
	}

	log.Verbose.Println("Packet processed.")

	// Determine the destination

	// Dispatch it to the correct channel

}
