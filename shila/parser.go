package shila

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func decodePacketID(p *Packet) error {

	if p.id == nil {
		p.id = new(PacketID)
	}

	if ip4v, tcp, err := decodeIPv4andTCPLayer(p.RawPayload()); err != nil {
		return err
	} else {
		p.id.Src.IP 	= ip4v.SrcIP
		p.id.Src.Port 	= int(tcp.SrcPort)
		p.id.Dst.IP 	= ip4v.DstIP
		p.id.Dst.Port 	= int(tcp.DstPort)
	}

	return nil
}

// Start slow but correct..
func decodeIPv4andTCPLayer(raw []byte) (layers.IPv4, layers.TCP, error) {

	ipv4 := layers.IPv4{}
	tcp  := layers.TCP{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ipv4, &tcp)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(raw, &decoded); err != nil {
		return ipv4, tcp, Error(fmt.Sprint("Could not decode IPv4/TCP layer", " - ", err.Error()))
	}

	return ipv4, tcp, nil
}

/*
func DecodeIPv4Options(p *shila.Packet) error {
	opts, err := layer.DecodeIPv4POptions(p.Payload.Decoded.IPv4Decoding)
	if err != nil {
		return Error(fmt.Sprint("Could not decode IPv4TCPPacket options", " - ", err.Error()))
	}
	p.Payload.Decoded.IPv4Options = opts
	return nil
}

func DecodeMPTCPOptions(p *shila.Packet) error {
	opts, err := layer.DecodeMPTCPOptions(p.Payload.Decoded.TCPDecoding)
	if err != nil {
		return Error(fmt.Sprint("Could not decode MPTCP options", " - ", err.Error()))
	}
	p.Payload.Decoded.MPTCPOptions = opts
	return nil
}*/


