package shila

import (
	"fmt"
	"github.com/google/gopacket/layers"
)

type Packet struct {
	IP  IP
	TCP TCP
}

type IP struct {
	Raw    []byte
	Parsed layers.IPv4
}

type TCP struct {
	Parsed layers.TCP
}

func NewPacketFromRawIP(raw []byte) *Packet {
	return &Packet{
		IP:  IP{raw, layers.IPv4{}},
		TCP: TCP{layers.TCP{}},
	}
}

func (p *Packet) String() string {
	return fmt.Sprint("Size of packet: ", len(p.IP.Raw), " Bytes.")
}
