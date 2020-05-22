package router

import (
"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
)

type Router struct {
	addressesFromToken 	 map[mptcp.EndpointToken]	shila.NetFlow
	addressesFromDstIPv4 map[shila.IPAddressPortKey]		shila.NetFlow
}


func New() *Router {
	return &Router{
		addressesFromToken: 	make(map[mptcp.EndpointToken]shila.NetFlow),
		addressesFromDstIPv4: 	make(map[shila.IPAddressPortKey]shila.NetFlow),
	}
}

func (m Router) RetrieveFromMPTCPEndpointToken(token mptcp.EndpointToken) (shila.NetFlow, bool) {
	packetHeader, ok := m.addressesFromToken[token]
	return packetHeader, ok
}

func (m Router) RetrieveFromIPAddressPortKey(key shila.IPAddressPortKey) (shila.NetFlow, bool) {
	packetHeader, ok := m.addressesFromDstIPv4[key]
	return packetHeader, ok
}

func (m Router) InsertFromIPAddressPortKey(key shila.IPAddressPortKey, srcAddr shila.NetworkAddress, dstAddr shila.NetworkAddress, path shila.NetworkPath) error {

	if _, ok := m.addressesFromDstIPv4[key]; ok {
		return Error(fmt.Sprint("Unable to insert routing entry for destination IPv4 {", key ,"}. - Entry already exists."))
	} else {
		m.addressesFromDstIPv4[key] = shila.NetFlow{srcAddr, path, dstAddr}
	}
	return nil
}

func (m Router) UpdateFromSynAckMpCapable(p *shila.Packet) error {
	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if err := m.insertFromMPTCPEndpointKey(key, p.Flow.NetFlow); err != nil {
				return Error(fmt.Sprint("Unable to insert MPTCP endpoint key. - ", err.Error()))
			}
		} else {
			return Error(fmt.Sprint("Error in fetching MPTCP endpoint key. - ", err.Error()))
		}
	}
	return Error(fmt.Sprint("Packet does not contain MPTCP endpoint key."))
}

func (m Router) insertFromMPTCPEndpointKey(key mptcp.EndpointKey, flow shila.NetFlow) error {
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

