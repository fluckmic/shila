package parser

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"shila/layer"
	"shila/log"
	"shila/shila"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

// Start slow but correct..
func DecodeIPv4andTCPLayer(p *shila.Packet) error {
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &p.IP.Parsed, &p.TCP.Parsed)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(p.IP.Raw, &decoded); err != nil {
		return Error(fmt.Sprint("Could not decode layer", " - ", err.Error()))
	}
	return nil
}

func DecodeMPTCPOptions(p *shila.Packet) error {
	for _, option := range p.TCP.Parsed.Options {
		if option.OptionType == layers.TCPOptionKind(layer.TCPOptionKindMPTCP) {
			opt := layer.MPTCPOption{}
			if err := opt.DecodeFromTCPOption(option); err != nil {
				return Error(fmt.Sprint("Could not parse TCP Option", " - ", err.Error()))
			}
			MPTCPOptions = append(MPTCPOptions, opt)
			log.Verbose.Print(opt.OptionSubtype)
		}
	}
	return nil
}

func parseTCPIPv4(p *shila.Packet) error {
	// Use the gopacket functionality to parse IPv4 and TCP layer

}

func parseIPOptions(p *shila.Packet) error {
	return nil
}
