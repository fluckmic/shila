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
	entryPoint Endpoint
	ipConnId   IPConnectionIdentifier
	netConnId  NetworkConnectionIdentifier
	payload    PacketPayload
}

// Has to be parsed for every packet
type IPConnectionIdentifier struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

func (iph *IPConnectionIdentifier) key() IPConnectionIdentifierKey {
	return KeyGenerator{}.IPConnectionIdentifierKey(*iph)
}

func (iph *IPConnectionIdentifier) srcKey() IPAddressPortKey {
	return KeyGenerator{}.IPAddressPortKey(iph.Src)
}

func (iph *IPConnectionIdentifier) srcIPKey() IPAddressKey {
	return KeyGenerator{}.IPAddressKey(iph.Src.IP)
}

func (iph *IPConnectionIdentifier) dstKey() IPAddressPortKey {
	return KeyGenerator{}.IPAddressPortKey(iph.Dst)
}

func (iph *IPConnectionIdentifier) dstIPKey() IPAddressKey {
	return KeyGenerator{}.IPAddressKey(iph.Dst.IP)
}

type NetworkConnectionIdentifier struct {
	Src  NetworkAddress
	Path NetworkPath
	Dst  NetworkAddress
}

func (nh *NetworkConnectionIdentifier) key() NetworkConnectionIdentifierKey {
	return KeyGenerator{}.NetworkConnectionIdentifierKey(*nh)
}

func (nh *NetworkConnectionIdentifier) destAndPathKey() NetworkAddressAndPathKey {
	return KeyGenerator{}.NetworkAddressAndPathKey(nh.Dst, nh.Path)
}

type IPv4TCPPacket struct {
	Raw      []byte
}

func NewPacket(ep Endpoint, iph IPConnectionIdentifier, raw []byte) *Packet {
	return &Packet{entryPoint: ep, ipConnId: iph, payload: PacketPayload{raw}}
}

func NewPacketInclNetConnId(ep Endpoint, iph IPConnectionIdentifier, nh NetworkConnectionIdentifier, raw []byte) *Packet {
	return &Packet{entryPoint: ep, ipConnId: iph, netConnId: nh, payload: PacketPayload{raw}}
}

func (p *Packet) GetIPConnId() IPConnectionIdentifier {
	return p.ipConnId
}

func (p *Packet) IPConnIdKey() IPConnectionIdentifierKey {
	return p.ipConnId.key()
}

func (p *Packet) IPConnIdDstKey() IPAddressPortKey {
	return p.ipConnId.dstKey()
}

func (p *Packet) IPConnIdSrcKey() IPAddressPortKey {
	return p.ipConnId.srcKey()
}

func (p *Packet) IPConnIdDstIPKey() IPAddressKey {
	return p.ipConnId.dstIPKey()
}

func (p *Packet) IPConnIdSrcIPKey() IPAddressKey {
	return p.ipConnId.srcIPKey()
}

func (p *Packet) GetNetworkConnId() NetworkConnectionIdentifier {
	return p.netConnId
}

func (p *Packet) SetNetworkConnId(netConnId NetworkConnectionIdentifier) {
	p.netConnId = netConnId
}

func (p *Packet) NetworkConnIdDstAndPathKey() NetworkAddressAndPathKey {
	 return p.netConnId.destAndPathKey()
}

func (p *Packet) GetRawPayload() []byte {
	return p.payload.Raw
}

func (p *Packet) GetEntryPoint() Endpoint {
	return p.entryPoint
}

func (p *Packet) PrintAllInfo() {
	log.Verbose.Print("########################################################")
	log.Verbose.Print("IP Header: ", KeyGenerator{}.IPConnectionIdentifierKey(p.ipConnId))
	// TODO: Network header --> nil
	log.Verbose.Print("Entry point: ", p.GetEntryPoint().Label())
	log.Verbose.Print("########################################################")
}