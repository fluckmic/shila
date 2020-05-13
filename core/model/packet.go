package model

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"shila/layer"
	"shila/log"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type PacketPayload IPv4TCPPacket

type Packet struct {
	entryPoint    Endpoint
	ipHeader      *IPHeader
	networkHeader *NetworkHeader
	payload       PacketPayload
}

// Has to be parsed for every packet
type IPHeader struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

// (src-ipv4-address:port<>dst-ipv4-address:port)
type IPHeaderKey   string
func (iph *IPHeader) key() IPHeaderKey {
	return IPHeaderKey(fmt.Sprint("(", iph.Src.String(), "<>", iph.Dst.String(), ")"))
}

// (src-ipv4-address:port)
func (iph *IPHeader) srcKey() IPAddressKey {
	return IPAddressKey(fmt.Sprint("(", iph.Src.String(),")"))
}

// (dst-ipv4-address:port)
func (iph *IPHeader) dstKey() IPAddressKey {
	return IPAddressKey(fmt.Sprint("(", iph.Dst.String(),")"))
}

type NetworkHeader struct {
	Src  NetworkAddress
	Path NetworkPath
	Dst  NetworkAddress
}

// (src-network-address<>path<>dst-network-address)
type NetworkHeaderKey string
func (nh *NetworkHeader) key() NetworkHeaderKey {
	return NetworkHeaderKey(fmt.Sprint("(",nh.Src.String(),"<>",nh.Path.String(),"<>",nh.Dst.String(),")"))
}

func (nh *NetworkHeader) destAndPathKey() NetworkAddressAndPathKey {
	return NetworkAddressAndPathKey(fmt.Sprint("(",nh.Dst.String(),"<>",nh.Path.String(),")"))
}

type IPv4TCPPacket struct {
	Raw      []byte
}

func NewPacketFromRawIP(ep Endpoint, raw []byte) *Packet {
	return &Packet{ep,nil, nil, PacketPayload{raw}}
}

func (p *Packet) GetIPHeader() (*IPHeader, error) {
	if p.ipHeader == nil {
		if err := p.decodeIPHeader(); err != nil {
			return nil, Error(fmt.Sprint("Unable to decode IP header. - ", err.Error()))
		}
	}
	return p.ipHeader, nil
}

func (p *Packet) IPHeaderKey() (IPHeaderKey, error) {
	if header, err := p.GetIPHeader(); err != nil {
		return "", err
	} else {
		return header.key(), nil
	}
}

func (p *Packet) IPHeaderDstKey() (IPAddressKey, error) {
	if header, err := p.GetIPHeader(); err != nil {
		return "", err
	} else {
		return header.dstKey(), nil
	}
}

func (p *Packet) IPHeaderSrcKey() (IPAddressKey, error) {
	if header, err := p.GetIPHeader(); err != nil {
		return "", err
	} else {
		return header.srcKey(), nil
	}
}

func (p *Packet) GetNetworkHeader(router *Mapping) (*NetworkHeader, error) {
	if p.networkHeader == nil {
		if err := p.decodeNetworkHeader(router); err != nil {
			return nil, err
		}
	}
	return p.networkHeader, nil
}

// Used by:
// - Receiving network endpoint
// - Sending established connection
func (p *Packet) SetNetworkHeader(header *NetworkHeader) {
	p.networkHeader = header
}

func (p *Packet) NetworkHeaderDstAndPathKey() (NetworkAddressAndPathKey, error) {
	if p.networkHeader == nil {
		// TODO: should not happen.
		panic("implement me.")
	} else {
		return p.networkHeader.destAndPathKey(), nil
	}
}

func (p *Packet) RawPayload() []byte {
	return p.payload.Raw
}

func (p *Packet) EntryPoint() Endpoint {
	return p.entryPoint
}

func (p *Packet) decodeIPHeader() error {

	if p.ipHeader == nil {
		p.ipHeader = new(IPHeader)
	}

	if ip4v, tcp, err := decodeIPv4andTCPLayer(p.RawPayload()); err != nil {
		return err
	} else {
		p.ipHeader.Src.IP 	= ip4v.SrcIP
		p.ipHeader.Src.Port 	= int(tcp.SrcPort)
		p.ipHeader.Dst.IP 	= ip4v.DstIP
		p.ipHeader.Dst.Port 	= int(tcp.DstPort)
	}

	return nil
}

func (p *Packet) decodeNetworkHeader(router *Mapping) error {

	// Decoding the packet networkHeader should just be necessary when a packet
	// for a new flow is received from a kernel endpoint.

	// If the MPTCP_JOIN option is set, then the packet is part of
	// a new subflow belonging to an already existing main flow.
	if token, MP_JOIN, err := p.getMPTCPReceiverToken(); MP_JOIN {
		if err != nil {
			// Retrieve address and path from the mapping
			if packetHeader, ok := router.RetrieveFromReceiverToken(token); ok {
				p.networkHeader = &packetHeader
				return nil
			} else {
				log.Info.Print("No routing entry found for packet {", p.ipHeader.key(), "} and token {", token, "}.")
			}
		} else {
			log.Info.Print("Error while fetching MPTCP receiver token for packet {", p.ipHeader.key(), "}.")
		}
	}

	// If the IP option contains valid shila options, then the
	// address and path can be fetched via these options
	if packetHeader, IP_OPTIONS, err := p.getNetworkHeaderFromIPOptions(); IP_OPTIONS {
		if err != nil {
			p.networkHeader = &packetHeader
			return nil
		} else {
			log.Info.Print("Error while fetching packet networkHeader from IP options for packet {", p.ipHeader.key(), "}.")
		}
	}

	// Last chance to fetch the destination address and path is
	// to do a lookup in the routing table using destination ip and
	// port as key.
	if packetHeader, ok := router.RetrieveFromIPAddressKey(p.ipHeader.dstKey()); ok {
		p.networkHeader = &packetHeader
		return nil
	} else {
		log.Info.Print("No routing entry found for packet {", p.ipHeader.key(), "} and key {", p.ipHeader.dstKey(), "}.")
	}

	return Error(fmt.Sprint("Unable to decode packet networkHeader for {", p.ipHeader.key(), "} - " +
		"No valid option to retrieve destination network address successful."))
}

func (p *Packet) getMPTCPReceiverToken() (MptcpReceiverToken, bool, error) {

	// We parse the IPv4 and the TCP layer again. Getting the receiver token is done
	// only once at the setup of a new subflow. It should be fine to do this twice.
	if _, tcp, err := decodeIPv4andTCPLayer(p.payload.Raw); err != nil {
		if mptcpOptions, err := layer.DecodeMPTCPOptions(tcp); err != nil {
			for _, mptcpOption := range mptcpOptions {
				if mptcpJoinOptionSYN, ok := mptcpOption.(layer.MPTCPJoinOptionSYN); ok {
					return MptcpReceiverToken(mptcpJoinOptionSYN.ReceiverToken), true, nil
				}
			}
			// MPTCP options does not contain the receiver token
			return MptcpReceiverToken(0), false, nil
		} else {
			// Error in decoding the mptcp options
			return MptcpReceiverToken(0), false, err
		}
	} else {
		// Error in decoding the ipv4/tcp options
		return MptcpReceiverToken(0), false, err
	}
}

func (p *Packet) getNetworkHeaderFromIPOptions() (NetworkHeader, bool, error) {
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
	return NetworkHeader{}, false, nil
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




