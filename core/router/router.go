package router

import (
	"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
)

type Router struct {
	mainIPFlows mapEndPointTokenToMainIPFlowKeys
	entries 	mapMainIPFlowKeyToRoutingEntries
}



type destinationMappings struct {
	fromMPTCPToken map[mptcp.EndpointToken]		shila.IPFlowKey
	fromIPPortKey  map[shila.IPAddressPortKey]	shila.NetworkAddress
}

type mapMainIPFlowKeyToRoutingEntries map[shila.IPFlowKey] *entry
type mapEndPointTokenToMainIPFlowKeys map[mptcp.EndpointToken] shila.IPFlowKey

type entry struct {
	dstAddr	 shila.NetworkAddress
	paths 	 []shila.NetworkPath
}

func New() Router {

	router := Router{
		entries: 		make(mapMainIPFlowKeyToRoutingEntries),
		destinations: destinationMappings{
			fromMPTCPToken: ),
			fromIPPortKey:  make(map[shila.IPAddressPortKey]	shila.NetworkAddress),
		},
	}

	// See whether there is some routing from it which can be loaded
	_ = router.fillWithEntriesFromDisk()
	return router
}

func (router *Router) Route(packet *shila.Packet) (Response, error) {

	// If the packet contains a receiver token, then the new connection is a MPTCP subflow flow.
	if token, ok, err := mptcp.GetReceiverToken(packet.Payload); ok {

	}

	// For a MPTCP MainFlow flow the network flow can probably be extracted from the IP options
	if netFlow, ok, err := router.getFromIPOptions(p.Payload); ok {
		if err == nil {
			return Response{ Dst: netFlow.Dst, From: IPOptions, IPFlow: p.Flow.IPFlow}, nil
		} else {
			return Response{}, shila.PrependError(err, "Unable to get IP options.")
		}
	}

	// For a MPTCP MainFlow flow the network flow is probably available in the router table
	if netFlow, ok := router.getFromIPAddressPortKey(p.Flow.IPFlow.DstKey()); ok {
		return  Response{ Dst: netFlow.Dst, From: RoutingTable, IPFlow: p.Flow.IPFlow}, nil
	}

	return Response{}, shila.TolerableError("No routing information available.")
}

func (router *Router) InsertDestinationFromIPAddressPortKey(key shila.IPAddressPortKey, dstAddr shila.NetworkAddress) error {
	if _, ok := router.destinations.fromIPPortKey[key]; ok {
		return shila.TolerableError("Response already exists.")
	} else {
		router.destinations.fromIPPortKey[key] = dstAddr
		return nil
	}
}

func (router *Router) InsertEndpointTokenToIPFlow(p *shila.Packet) error {
	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return shila.PrependError(err, fmt.Sprint("Unable to convert token from key."))
			} else {
				if _, ok := router.destinations.fromMPTCPToken[token]; ok {
					return shila.TolerableError("Response already exists.")
				} else {
					router.destinations.fromMPTCPToken[token] = p.Flow.IPFlow.Key()
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

func (router *Router) Says(str string) string {
	return  fmt.Sprint(router.Identifier(), ": ", str)
}

func (router *Router) Identifier() string {
	return fmt.Sprint("Router")
}

func (router *Router) fillWithEntriesFromDisk() error {
	routingEntries, err := loadRoutingEntriesFromDisk()
	if err != nil {
		return PrependError(err, "Unable to load entries from disk.")
	}
	err = router.batchInsert(routingEntries)
	return nil
}

if err == nil {
flow, ok := router.getFromMPTCPEndpointToken(token)
if ok {
return Response{ Dst: flow.NetFlow.Dst, From: MPTCPEndpointToken, IPFlow: flow.IPFlow }, nil
} else {
return Response{}, shila.TolerableError(fmt.Sprint("No network flow for MPTCP receiver token ", token, "."))
}
} else {
return Response{}, shila.PrependError(err, "Unable to fetch MPTCP receiver token.")
}