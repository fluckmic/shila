package model

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"shila/layer"
	"shila/log"
	"shila/routing"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type PacketPayload IPv4TCPPacket

type Key_SrcIPv4DstIPv4_ string		/* (1.2.3.4:23<>5.6.7.8:45) */
type Key_DstIPv4_	 	 string 	/* (5.6.7.8:45) 			*/

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

func (p *PacketID) key() Key_SrcIPv4DstIPv4_ {
	return Key_SrcIPv4DstIPv4_(fmt.Sprint("(", p.Src.String(), "<>", p.Dst.String(), ")"))
}

type IPv4TCPPacket struct {
	Raw      []byte
}

func NewPacketFromRawIP(ep Endpoint, raw []byte) *Packet {
	return &Packet{ep,nil, nil, PacketPayload{raw}}
}

func (p *Packet) ID() (*PacketID, error) {
	if p.id == nil {
		if err := p.decodePacketID(); err != nil {
			return nil, Error(fmt.Sprint("Could not decode packet id - ", err.Error()))
		}
	}
	return p.id, nil
}

func (p *Packet) Key() (Key_SrcIPv4DstIPv4_, error) {
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
		if err := p.decodePacketHeader(); err != nil {
			return nil, err
		}
	}
	return p.header, nil
}

// Used by:
// - Receiving network endpoint
// - Sending established connection
func (p *Packet) SetPacketHeader(header *PacketHeader) {
	p.header = header
}

func(p *Packet) decodePacketID() error {

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

func (p *Packet) decodePacketHeader() error {

	key, err := p.Key()
	if err != nil {
		return Error(fmt.Sprint("Unable to decode packet header. - ", err.Error()))
	}

	// Decoding the packet header should just be necessary when a packet
	// for a new flow is received from a kernel endpoint.

	// If the MPTCP_JOIN option is set, then the packet is part of
	// a new subflow belonging to an already existing main flow.
	if token, MP_JOIN, err := p.getMPTCPReceiverToken(); MP_JOIN {
		if err != nil {
			// Retrieve address and path from the mapping
			_ = token

		} else {
			log.Info.Print("Error while fetching MPTCP receiver token for {", key, "}.")
		}
	}

	// If the IP option contains valid shila options, then the
	// address and path can be fetched via these options
	if packetHeader, IP_OPTIONS, err := p.getPacketHeaderFromIPOptions(); IP_OPTIONS {
		if err != nil {
			p.header = packetHeader
		} else {
			log.Info.Print("Error while fetching packet header from IP options for {", key, "}.")
		}
	}

	// Last chance to fetch the destination address and path is
	// to do a lookup in the routing table using destination ip and
	// port as key.
	destKey := Key_DstIPv4_(p.id.Dst.String())
	_ = destKey

	return Error(fmt.Sprint("Unable to decode packet header for {", key, "} - " +
		"No valid option to retrieve destination network address successful."))
}

// Start slow but correct..
func decodeIPv4andTCPLayer(raw []byte) (layers.IPv4, layers.TCP, error) {

	ipv4 := layers.IPv4{}
	tcp  := layers.TCP{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ipv4, &tcp)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(raw, &decoded); err != nil {
		return ipv4, tcp, Error(fmt.Sprint("Could not decode IPv4/TCP layer. - ", err.Error()))
	}

	return ipv4, tcp, nil
}

func (p *Packet) getMPTCPReceiverToken() (routing.Key_MPTCPReceiverToken_, bool, error) {

	// We parse the IPv4 and the TCP layer again. Getting the receiver token is done
	// only once at the setup of a new subflow. It should be fine to do this twice.
	if _, tcp, err := decodeIPv4andTCPLayer(p.payload.Raw); err != nil {
		if mptcpOptions, err := layer.DecodeMPTCPOptions(tcp); err != nil {
			for _, mptcpOption := range mptcpOptions {
				if mptcpJoinOptionSYN, ok := mptcpOption.(layer.MPTCPJoinOptionSYN); ok {
					return routing.Key_MPTCPReceiverToken_(mptcpJoinOptionSYN.ReceiverToken), true, nil
				}
			}
			// MPTCP options does not contain the receiver token
			return routing.Key_MPTCPReceiverToken_(0), false, nil
		} else {
			// Error in decoding the mptcp options
			return routing.Key_MPTCPReceiverToken_(0), false, err
		}
	} else {
		// Error in decoding the ipv4/tcp options
		return routing.Key_MPTCPReceiverToken_(0), false, err
	}
}

func (p *Packet) getPacketHeaderFromIPOptions() (*PacketHeader, bool, error) {
	// TODO!
	/*
	func DecodeIPv4Options(p *shila.Packet) error {
		opts, err := layer.DecodeIPv4POptions(p.Payload.Decoded.IPv4Decoding)
		if err != nil {
			return Error(fmt.Sprint("Could not decode IPv4TCPPacket options", " - ", err.Error()))
		}
		p.Payload.Decoded.IPv4Options = opts
		return nil
	}
	*/
	return nil, false, nil
}




