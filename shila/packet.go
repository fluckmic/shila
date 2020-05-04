package shila

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type PacketPayload  IPv4TCPPacket

type Packet struct {
	entryPoint Endpoint
	id         *PacketID
	header     *PacketHeader
	payload    PacketPayload
}

// Has to be parsed for every packet
type PacketID struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

type PacketHeader struct {
	Src  NetworkAddress
	Path NetworkPath
	Dst  NetworkAddress
}

func (p *PacketID) key() string {
	return fmt.Sprint("{", p.Src.String(), "<>", p.Dst.String(), "}")
}

type IPv4TCPPacket struct {
	Raw      []byte
}

func NewPacketFromRawIP(ep Endpoint, raw []byte) *Packet {
	return &Packet{ep,nil, nil, PacketPayload{raw}}
}

func (p *Packet) ID() (*PacketID, error) {
	if p.id == nil {
		if err := decodePacketID(p); err != nil {
			return nil, Error(fmt.Sprint("Could not decode packet id - ", err.Error()))
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

func (p *Packet) EntryPoint() Endpoint {
	return p.entryPoint
}

func (p *Packet) PacketHeader() (*PacketHeader, error) {
	if p.header == nil {
		if err := decodePacketHeader(p); err != nil {
			return nil, Error(fmt.Sprint("Could not decode packet header - ", err.Error()))
		}
	}
	return p.header, nil
}

func (p *Packet) SetPacketHeader(header *PacketHeader) {
	p.header = header
}

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

func decodePacketHeader(p *Packet) error {
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
