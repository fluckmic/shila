package shila

import (
	"fmt"
	"github.com/google/gopacket/layers"
	"shila/layer"
)

type Packet struct {
	IP  IP
	TCP TCP
}

type IP struct {
	Raw     []byte
	Parsed  layers.IPv4
	Options []layer.IPv4Option
}

type TCP struct {
	Parsed       layers.TCP
	MPTCPOptions []layer.MPTCPOption
}

func NewPacketFromRawIP(raw []byte) *Packet {
	return &Packet{
		IP:  IP{raw, layers.IPv4{}, []layer.IPv4Option{}},
		TCP: TCP{layers.TCP{}, []layer.MPTCPOption{}},
	}
}

func (p *Packet) String() string {
	// TODO: Debug...
	return fmt.Sprint("Size of packet: ", len(p.IP.Raw), " Bytes.")
}
