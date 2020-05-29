//
package netflow

import (
	"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
	"shila/log"
)

type Router struct {
	flows mappings
}

type mappings struct {
	fromMPTCPToken map[mptcp.EndpointToken]		shila.NetFlow
	fromIPPortKey  map[shila.IPAddressPortKey]	shila.NetFlow
}

func NewRouter() Router {

	router := Router{
		flows: mappings{
			fromMPTCPToken: make(map[mptcp.EndpointToken]shila.NetFlow),
			fromIPPortKey:  make(map[shila.IPAddressPortKey]shila.NetFlow),
		},
	}

	// See whether there is some routing from it which can be loaded
	if err := router.fillWithEntriesFromDisk(); err != nil {
		log.Info.Print("Unable to load routing entries from disk. - ", err.Error())
	}

	return router
}

func (r *Router) InsertFromIPAddressPortKey(key shila.IPAddressPortKey, flow shila.NetFlow) error {
	if _, ok := r.flows.fromIPPortKey[key]; ok {
		return shila.TolerableError("Entry already exists.")
	} else {
		r.flows.fromIPPortKey[key] = flow
		return nil
	}
}

func (r *Router) InsertFromSynAckMpCapable(p *shila.Packet, flow shila.NetFlow) error {
	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return shila.PrependError(err, fmt.Sprint("Unable to convert token from key."))
			} else {
				if _, ok := r.flows.fromMPTCPToken[token]; ok {
					return shila.TolerableError("Entry already exists.")
				} else {
					r.flows.fromMPTCPToken[token] = flow
					return nil
				}
			}
		} else {
			return shila.PrependError(err, "Unable to fetch MPTCP endpoint key.")
		}
	} else {
		// The packet does not necessarily contain the endpoint key
		// (e.g. for a packet belonging to a subflow)
		return nil
	}
}

func (r *Router) Route(p *shila.Packet) (shila.NetFlow, shila.FlowKind, error) {

	// If the packet contains a receiver token, then the new connection is a MPTCP subflow flow.
	if token, ok, err := mptcp.GetReceiverToken(p.Payload); ok {
		if err == nil {
			if netFlow, ok := r.getFromMPTCPEndpointToken(token); ok {
				return netFlow, shila.Subflow, nil
			} else {
				return shila.NetFlow{}, shila.Unknown,
				shila.TolerableError(fmt.Sprint("No network flow for MPTCP receiver token {", token, "}."))
			}
		} else {
			return shila.NetFlow{}, shila.Unknown, shila.PrependError(err, "Unable to fetch MPTCP receiver token.")
		}
	}

	// For a MPTCP Mainflow flow the network flow can probably be extracted from the IP options
	if netFlow, ok, err := r.getFromIPOptions(p.Payload); ok {
		if err == nil {
			return netFlow, shila.Mainflow, nil
		} else {
			return shila.NetFlow{}, shila.Unknown, shila.PrependError(err, "Unable to get IP options.")
		}
	}

	// For a MPTCP Mainflow flow the network flow is probably available in the router table
	if netFlow, ok := r.getFromIPAddressPortKey(p.Flow.IPFlow.DstKey()); ok {
		return netFlow, shila.Mainflow, nil
	}

	return shila.NetFlow{}, shila.Unknown, shila.TolerableError("No routing information available.")
}

func (r *Router) getFromIPOptions(raw []byte) (shila.NetFlow, bool, error) {
	// TODO: https://github.com/fluckmic/shila/issues/17
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

func (r *Router) fillWithEntriesFromDisk() error {
	routingEntries, err := loadRoutingEntriesFromDisk()
	if err != nil {
		return err
	}
	err = r.batchInsert(routingEntries)
	return nil
}