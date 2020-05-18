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
	entryPoint        Endpoint
	ipConnection      IPConnectionTuple
	networkConnection NetworkConnectionTriple
	payload           PacketPayload
}

// Has to be parsed for every packet
type IPConnectionTuple struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

func (iph *IPConnectionTuple) key() IPConnectionTupleKey {
	return KeyGenerator{}.IPConnectionTupleKey(*iph)
}

func (iph *IPConnectionTuple) srcKey() IPAddressPortKey {
	return KeyGenerator{}.IPAddressPortKey(iph.Src)
}

func (iph *IPConnectionTuple) srcIPKey() IPAddressKey {
	return KeyGenerator{}.IPAddressKey(iph.Src.IP)
}

func (iph *IPConnectionTuple) dstKey() IPAddressPortKey {
	return KeyGenerator{}.IPAddressPortKey(iph.Dst)
}

func (iph *IPConnectionTuple) dstIPKey() IPAddressKey {
	return KeyGenerator{}.IPAddressKey(iph.Dst.IP)
}

type NetworkConnectionTriple struct {
	Src  NetworkAddress
	Path NetworkPath
	Dst  NetworkAddress
}

func (nh *NetworkConnectionTriple) key() NetworkConnectionTripleKey {
	return KeyGenerator{}.NetworkConnectionTripleKey(*nh)
}

func (nh *NetworkConnectionTriple) destAndPathKey() NetworkAddressAndPathKey {
	return KeyGenerator{}.NetworkAddressAndPathKey(nh.Dst, nh.Path)
}

type IPv4TCPPacket struct {
	Raw      []byte
}

func NewPacket(ep Endpoint, iph IPConnectionTuple, raw []byte) *Packet {
	return &Packet{entryPoint: ep, ipConnection: iph, payload: PacketPayload{raw}}
}

func NewPacketInclNetworkHeader(ep Endpoint, iph IPConnectionTuple, nh NetworkConnectionTriple, raw []byte) *Packet {
	return &Packet{entryPoint: ep, ipConnection: iph, networkConnection: nh, payload: PacketPayload{raw}}
}

func (p *Packet) GetIPHeader() IPConnectionTuple {
	return p.ipConnection
}

func (p *Packet) IPHeaderKey() IPConnectionTupleKey {
	return p.ipConnection.key()
}

func (p *Packet) IPHeaderDstKey() IPAddressPortKey {
	return p.ipConnection.dstKey()
}

func (p *Packet) IPHeaderSrcKey() IPAddressPortKey {
	return p.ipConnection.srcKey()
}

func (p *Packet) IPHeaderDstIPKey() IPAddressKey {
	return p.ipConnection.dstIPKey()
}

func (p *Packet) IPHeaderSrcIPKey() IPAddressKey {
	return p.ipConnection.srcIPKey()
}

func (p *Packet) GetNetworkHeader() NetworkConnectionTriple {
	return p.networkConnection
}

func (p *Packet) SetNetworkHeader(header NetworkConnectionTriple) {
	p.networkConnection = header
}

func (p *Packet) NetworkHeaderDstAndPathKey() NetworkAddressAndPathKey {
	 return p.networkConnection.destAndPathKey()
}

func (p *Packet) GetRawPayload() []byte {
	return p.payload.Raw
}

func (p *Packet) GetEntryPoint() Endpoint {
	return p.entryPoint
}

func (p *Packet) PrintAllInfo() {
	log.Verbose.Print("########################################################")
	log.Verbose.Print("IP Header: ", KeyGenerator{}.IPConnectionTupleKey(p.ipConnection))
	// TODO: Network header --> nil
	log.Verbose.Print("Entry point: ", p.GetEntryPoint().Label())
	log.Verbose.Print("########################################################")
}