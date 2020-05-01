package shila

import (
	"fmt"
	"net"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type PacketPayload  IPv4TCPPacket

// TODO: Add an additional label to make parsing of entry point easier.

type Packet struct {
	entryPoint  endpoint
	id			*PacketID
	header 		*PacketHeader
	payload 	PacketPayload
}

// Has to be parsed for every packet
type PacketID struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

func (p *PacketID) key() string {
	return fmt.Sprint("{", p.Src.String(), "<>", p.Dst.String(), "}")
}

type PacketHeader struct {
}

type IPv4TCPPacket struct {
	Raw      []byte
}

func NewPacketFromRawIP(ep endpoint, raw []byte) *Packet {
	return &Packet{ep,nil, nil, PacketPayload{raw}}
}

func (p *Packet) ID() (*PacketID, error) {
	if p.id == nil {
		if err := decodePacketID(p); err != nil {
			return nil, Error(fmt.Sprint("Could not decode packet id", " - ", err.Error()))
		}
	}
	return p.id, nil
}

func (p *Packet) Key() (string, error) {
	if key, err := p.ID(); err != nil {
		return "", err
	} else {
		return key.key(), nil
	}
}

func (p *Packet) RawPayload() []byte {
	return p.payload.Raw
}

func (p *Packet) EntryPoint() endpoint {
	return p.entryPoint
}