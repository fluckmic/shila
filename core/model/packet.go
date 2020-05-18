package model

import (
	"net"
	"shila/log"
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

func (iph *IPHeader) key() IPHeaderKey {
	return KeyGenerator{}.IPHeaderKey(*iph)
}

func (iph *IPHeader) srcKey() IPAddressPortKey {
	return KeyGenerator{}.IPAddressPortKey(iph.Src)
}

func (iph *IPHeader) srcIPKey() IPAddressKey {
	return KeyGenerator{}.IPAddressKey(iph.Src.IP)
}

func (iph *IPHeader) dstKey() IPAddressPortKey {
	return KeyGenerator{}.IPAddressPortKey(iph.Dst)
}

func (iph *IPHeader) dstIPKey() IPAddressKey {
	return KeyGenerator{}.IPAddressKey(iph.Dst.IP)
}

type NetworkHeader struct {
	Src  NetworkAddress
	Path NetworkPath
	Dst  NetworkAddress
}

func (nh *NetworkHeader) key() NetworkHeaderKey {
	return KeyGenerator{}.NetworkHeaderKey(*nh)
}

func (nh *NetworkHeader) destAndPathKey() NetworkAddressAndPathKey {
	return KeyGenerator{}.NetworkAddressAndPathKey(nh.Dst, nh.Path)
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

func (p *Packet) IPHeaderDstKey() IPAddressPortKey {
	return p.ipHeader.dstKey()
}

func (p *Packet) IPHeaderSrcKey() IPAddressPortKey {
	return p.ipHeader.srcKey()
}

func (p *Packet) IPHeaderDstIPKey() IPAddressKey {
	return p.ipHeader.dstIPKey()
}

func (p *Packet) IPHeaderSrcIPKey() IPAddressKey {
	return p.ipHeader.srcIPKey()
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

func (p *Packet) PrintAllInfo() {
	log.Verbose.Print("########################################################")
	log.Verbose.Print("IP Header: ", KeyGenerator{}.IPHeaderKey(p.ipHeader))
	// TODO: Network header --> nil
	log.Verbose.Print("Entry point: ", p.GetEntryPoint().Label())
	log.Verbose.Print("########################################################")
}