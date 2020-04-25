package parser

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"shila/layer"
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
		return Error(fmt.Sprint("Could not decode IPv4/TCP layer", " - ", err.Error()))
	}
	return nil
}

func DecodeIPv4Options(p *shila.Packet) error {
	opts, err := layer.DecodeIPv4POptions(p.IP.Parsed)
	if err != nil {
		return Error(fmt.Sprint("Could not decode IP options", " - ", err.Error()))
	}
	p.IP.Options = opts
	return nil
}

func DecodeMPTCPOptions(p *shila.Packet) error {
	opts, err := layer.DecodeMPTCPOptions(p.TCP.Parsed)
	if err != nil {
		return Error(fmt.Sprint("Could not decode MPTCP options", " - ", err.Error()))
	}
	p.TCP.MPTCPOptions = opts
	return nil
}
