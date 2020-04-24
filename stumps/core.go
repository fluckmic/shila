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

	// Parse the packet
	if err := parser.Parse(p); err != nil {
		log.Error.Panicln(err.Error())
	}

	// Determine the destination

	// Dispatch it to the correct channel

}
