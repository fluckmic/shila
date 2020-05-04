package stumps

import (
	"shila/kernelSide"
	"shila/log"
	"shila/shila"
	"shila/shila/connection"
)

// Will be part of the core config
var nKerepIngressHandler = 2

var kernsi	 *kernelSide.Manager
var mapping  *connection.Mapping

func Setup(kersi *kernelSide.Manager) error {
	kernsi = kersi
	mapping 	= connection.NewMapping(kernsi, nil)
	return nil
}

func Start() error {
	for kerepKey, ep := range kernsi.Endpoints {
		go serveKerepIngress(ep.TrafficChannels().Ingress, string(kerepKey))
	}
	return nil
}

func serveKerepIngress(buffer chan *shila.Packet, kerepKey string) {
	for id := 0; id < nKerepIngressHandler; id++ {
		go handleKerepIngress(buffer, kerepKey, id)
	}
}

func handleKerepIngress(buffer chan *shila.Packet, kerepKey string, handlerId int) {
	//log.Verbose.Print("Started kernel endpoint ingress listener ", kerepKey, "-", handlerId, ".")
	for p := range buffer {
		processKerepIngress(p)
	}
}

func processKerepIngress(p *shila.Packet) {

	// Get the connection
	var con *connection.Connection
	if key, err := p.Key(); err != nil {
		log.Error.Panicln(err.Error())
	} else {
		con = mapping.Retrieve(connection.ID(key))
	}

	if err := con.ProcessPacket(p); err != nil {
		log.Error.Panicln(err.Error())
	}

	/*
	// Decode the IPv4 options of the packet
	if err := parser.DecodeIPv4Options(p); err != nil {
		log.Error.Panicln(err.Error())
	}

	// Decode the MPTCP options of the packet
	if err := parser.DecodeMPTCPOptions(p); err != nil {
		log.Error.Panicln(err.Error())
	}
	 */

	// Determine the destination

	// Dispatch it to the correct channel

}
