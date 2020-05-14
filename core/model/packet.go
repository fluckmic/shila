package model

import (
	"fmt"
	"net"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type PacketPayload IPv4TCPPacket

type Packet struct {
	entryPoint    Endpoint
	ipHeader      IPHeader
	networkHeader NetworkHeader
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

func NewPacket(ep Endpoint, iph IPHeader, raw []byte) *Packet {
	return &Packet{entryPoint: ep, ipHeader: iph, payload: PacketPayload{raw}}
}

func NewPacketInclNetworkHeader(ep Endpoint, iph IPHeader, nh NetworkHeader, raw []byte) *Packet {
	return &Packet{entryPoint: ep, ipHeader: iph, networkHeader: nh, payload: PacketPayload{raw}}
}

func (p *Packet) GetIPHeader() IPHeader {
	return p.ipHeader
}

func (p *Packet) IPHeaderKey() IPHeaderKey {
	return p.ipHeader.key()
}

func (p *Packet) IPHeaderDstKey() IPAddressKey {
	return p.ipHeader.dstKey()
}

func (p *Packet) IPHeaderSrcKey() IPAddressKey {
	return p.ipHeader.srcKey()
}

func (p *Packet) GetNetworkHeader() NetworkHeader {
	return p.networkHeader
}

func (p *Packet) SetNetworkHeader(header NetworkHeader) {
	p.networkHeader = header
}

func (p *Packet) NetworkHeaderDstAndPathKey() NetworkAddressAndPathKey {
	 return p.networkHeader.destAndPathKey()
}

func (p *Packet) GetRawPayload() []byte {
	return p.payload.Raw
}

func (p *Packet) GetEntryPoint() Endpoint {
	return p.entryPoint
}