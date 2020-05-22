package shila

import (
	"fmt"
	"shila/layer/mptcp"
)

type Mapping struct {
	addressesFromToken 	 map[mptcp.EndpointToken]	NetFlow
	addressesFromDstIPv4 map[IPAddressPortKey]		NetFlow
}

func NewMapping() *Mapping {
	return &Mapping{make(map[mptcp.EndpointToken]NetFlow),
				   make(map[IPAddressPortKey]NetFlow)}
}

func (m Mapping) RetrieveFromMPTCPEndpointToken(token mptcp.EndpointToken) (NetFlow, bool) {
	packetHeader, ok := m.addressesFromToken[token]
	return packetHeader, ok
}

func (m Mapping) InsertFromMPTCPEndpointKey(key mptcp.EndpointKey, flow NetFlow) error {
	if token, err := mptcp.EndpointKeyToToken(key); err != nil {
		return Error(fmt.Sprint("Unable to convert from key to token. - ", err.Error()))
	} else {
		if _, ok := m.addressesFromToken[token]; ok {
			return Error(fmt.Sprint("Entry already exists."))
		} else {
			m.addressesFromToken[token] = flow
		}
	}
	return nil
}

func (m Mapping) RetrieveFromIPAddressPortKey(key IPAddressPortKey) (NetFlow, bool) {
	packetHeader, ok := m.addressesFromDstIPv4[key]
	return packetHeader, ok
}

func (m Mapping) InsertFromIPAddressPortKey(key IPAddressPortKey, srcAddr NetworkAddress, dstAddr NetworkAddress, path NetworkPath) error {

	if _, ok := m.addressesFromDstIPv4[key]; ok {
		return Error(fmt.Sprint("Unable to insert routing entry for destination IPv4 {", key ,"}. - Entry already exists."))
	} else {
		m.addressesFromDstIPv4[key] = NetFlow{srcAddr, path, dstAddr}
	}
	return nil
}
