package parser

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"shila/shila"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

// Start slow but correct..
// Also thinkable: Just extract the parameters needed instead of carrying
// around all the parsed parameters.
func Parse(p *shila.Packet) error {

	// First parse IPv4 and TCP
	if err := parseTCPIPv4(p); err != nil {
		return Error(fmt.Sprint("Unable to parse packet", " - ", err.Error()))
	}

	// Potentially parse the IP options
	if err := parseIPOptions(p); err != nil {
		return Error(fmt.Sprint("Unable to parse packet", " - ", err.Error()))
	}

	// Potentially parse the TCP options (MPTCP parameters)
	if err := parseTCPOptions(p); err != nil {
		return Error(fmt.Sprint("Unable to parse packet", " - ", err.Error()))
	}

	return nil
}

func parseTCPIPv4(p *shila.Packet) error {
	// Use the gopacket functionality to parse IPv4 and TCP layer
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &p.IP.Parsed, &p.TCP.Parsed)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(p.IP.Raw, &decoded); err != nil {
		return Error(fmt.Sprint("Could not decode layers", " - ", err.Error()))
	}
	return nil
}

func parseIPOptions(p *shila.Packet) error {
	return nil
}

func parseTCPOptions(p *shila.Packet) error {
	return nil
}
