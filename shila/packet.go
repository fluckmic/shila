package shila

import (
	"fmt"
	"github.com/google/gopacket/layers"
	"shila/layer"
	"shila/parser"
)

type PacketPayload  IPv4TCPPacket

type Packet struct {
	id 		*PacketID
	Payload PacketPayload
}

type PacketID struct {
	string
}

type IPv4TCPPacket struct {
	Raw      []byte
	Decoded  *IPv4TCPPacketDecoding
}

type IPv4TCPPacketDecoding struct {
	IPv4Decoding layers.IPv4
	IPv4Options  []layer.IPv4Option
	TCPDecoding  layers.TCP
	MPTCPOptions []layer.MPTCPOption
}

func NewPacketFromRawIP(raw []byte) *Packet {
	return &Packet{nil, PacketPayload{raw, nil}}
}

func (p *Packet) String() string {
	// TODO: Debug...
	return fmt.Sprint("Size of packet: ", len(p.Payload.Raw), " Bytes.")
}

func (p *Packet) ID() string {

	if p.id != nil {

		p.Payload.Decoded = &IPv4TCPPacketDecoding{
			IPv4Decoding: layers.IPv4{},
			IPv4Options:  nil,
			TCPDecoding:  layers.TCP{},
			MPTCPOptions: nil,
		}

		// Decode the IPv4 and the TCP layer of the packet
		if err := parser.DecodeIPv4andTCPLayer(p); err != nil {
			log.Error.Panicln(err.Error())
		}
	} else {
		return p.id.string
	}


	return ""
}