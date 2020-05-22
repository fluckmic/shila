package netflow

import (
"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
)

type Router struct {
	flows mappings
}

type mappings struct {
	fromMPTCPToken map[mptcp.EndpointToken]		shila.NetFlow
	fromIPPortKey  map[shila.IPAddressPortKey]	shila.NetFlow
}

func NewRouter() *Router {
	return &Router{
		flows: mappings{
			fromMPTCPToken: make(map[mptcp.EndpointToken]shila.NetFlow),
			fromIPPortKey:  make(map[shila.IPAddressPortKey]shila.NetFlow),
		},
	}
}

func (r *Router) InsertFromIPAddressPortKey(key shila.IPAddressPortKey, flow shila.NetFlow) error {
	if _, ok := r.flows.fromIPPortKey[key]; ok {
		return Error(fmt.Sprint("Entry already exists."))
	} else {
		r.flows.fromIPPortKey[key] = flow
		return nil
	}
}

func (r *Router) InsertFromSynAckMpCapable(p *shila.Packet, flow shila.NetFlow) error {
	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return Error(fmt.Sprint("Unable to convert token from key. - ", err.Error()))
			} else {
				if _, ok := r.flows.fromMPTCPToken[token]; ok {
					return Error(fmt.Sprint("Entry already exists."))
				} else {
					r.flows.fromMPTCPToken[token] = flow
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

func (r *Router) Route(p *shila.Packet) (shila.NetFlow, shila.FlowKind, error) {

	// If the packet contains a receiver token, then the new connection is a MPTCP subflow flow.
	if token, ok, err := mptcp.GetReceiverToken(p.Payload); ok {
		if err == nil {
			if netFlow, ok := r.getFromMPTCPEndpointToken(token); ok {
				return netFlow, shila.Mainflow, nil
			} else {
				return shila.NetFlow{}, shila.Unknown,
				Error(fmt.Sprint("No network flow for MPTCP receiver token {", token, "}" +
					" beside having the packet containing it."))
			}
		} else {
			return shila.NetFlow{}, shila.Unknown,
			Error(fmt.Sprint("Unable to fetch MPTCP receiver token. - ", err.Error()))
		}
	}

	// For a MPTCP Mainflow flow the network flow can probably be extracted from the IP options
	if netFlow, ok, err := r.getFromIPOptions(p.Payload); ok {
		if err == nil {
			return netFlow, shila.Subflow, nil
		} else {
			return shila.NetFlow{}, shila.Unknown,
				Error(fmt.Sprint("Unable to get IP options. - ", err.Error()))
		}
	}

	// For a MPTCP Mainflow flow the network flow is probably available in the router table
	if netFlow, ok := r.getFromIPAddressPortKey(p.Flow.IPFlow.DstKey()); ok {
		return netFlow, shila.Mainflow, nil
	}

	return shila.NetFlow{}, shila.Unknown, Error(fmt.Sprint("No routing information available."))
}

func (r *Router) getFromIPOptions(raw []byte) (shila.NetFlow, bool, error) {
	return shila.NetFlow{}, false, nil
}

func (r *Router) getFromIPAddressPortKey(key shila.IPAddressPortKey) (shila.NetFlow, bool) {
	packetHeader, ok := r.flows.fromIPPortKey[key]
	return packetHeader, ok
}

func (r *Router) getFromMPTCPEndpointToken(token mptcp.EndpointToken) (shila.NetFlow, bool) {
	packetHeader, ok := r.flows.fromMPTCPToken[token]
	return packetHeader, ok
}
