package netflow

import (
"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
)

type Router struct {
	addressesFromToken 	 map[mptcp.EndpointToken]	shila.NetFlow
	addressesFromDstIPv4 map[shila.IPAddressPortKey]		shila.NetFlow
}

func NewRouter() *Router {
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

func (m Router) InsertFromIPAddressPortKey(key shila.IPAddressPortKey, flow shila.NetFlow) error {
	if _, ok := m.addressesFromDstIPv4[key]; ok {
		return Error(fmt.Sprint("Entry already exists."))
	} else {
		m.addressesFromDstIPv4[key] = flow
		return nil
	}
}

func (m Router) InsertFromSynAckMpCapable(p *shila.Packet, flow shila.NetFlow) error {
	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return Error(fmt.Sprint("Unable to convert token from key. - ", err.Error()))
			} else {
				if _, ok := m.addressesFromToken[token]; ok {
					return Error(fmt.Sprint("Entry already exists."))
				} else {
					m.addressesFromToken[token] = flow
					return nil
				}
			}
		} else {
			return Error(fmt.Sprint("Error in fetching MPTCP endpoint key. - ", err.Error()))
		}
	} else {
		// The packet does not necessarily contain the endpoint key (e.g. for a packet belonging to a subflow)
		return nil
	}
}


