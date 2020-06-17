package router

import (
	"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
)

type Router struct {
	mainIPFlows mapEndPointTokenToMainIPFlowKeys
	entries 	mapMainIPFlowKeyToRoutingEntries
	fixedTable	mapIPAddressPortKeyToDstAddresses
}

type mapMainIPFlowKeyToRoutingEntries map[shila.IPFlowKey] *Entry
type mapEndPointTokenToMainIPFlowKeys map[mptcp.EndpointToken] shila.IPFlowKey
type mapIPAddressPortKeyToDstAddresses map[shila.IPAddressPortKey] shila.NetworkAddress

func New() Router {

	router := Router{
		mainIPFlows:	make(mapEndPointTokenToMainIPFlowKeys),
		entries: 		make(mapMainIPFlowKeyToRoutingEntries),
		fixedTable:		make(mapIPAddressPortKeyToDstAddresses),
	}

	// See whether there is some routing from it which can be loaded
	_ = router.fillWithEntriesFromDisk()
	return router
}

func (router *Router) Route(packet *shila.Packet) (Response, error) {

	if mainIPFlowKey, ok := router.getMainIPFlowKeyFromEndpointToken(packet); ok {
		return router.routeSubFlow(mainIPFlowKey)
	}

	return router.routeMainFlow(packet)
}

func (router *Router) routeMainFlow(packet *shila.Packet) (Response, error) {

	// The key we get directly from the packet
	mainIPFlowKey := packet.Flow.IPFlow.Key()

	// For the destination we have to options:
	var dstAddr shila.NetworkAddress; var ok bool

	// 1) From the IP Options
	dstAddr, ok = router.getDestinationFromIPOptions(packet)
	if !ok {
		// 2) From the routing table
		dstAddr, ok = router.getDestinationFromIPAddressPortKey(packet)
	}

	if ok {
		entry := router.insertAndReturnRoutingEntry(mainIPFlowKey, dstAddr)
		return Response{
			Dst: 			entry.Dst,
			FlowCategory: 	MainFlow,
			MainIPFlowKey: 	mainIPFlowKey,
		},nil
	}

	return Response{}, shila.TolerableError("Unable to route packet. No routing information available.")
}

func (router *Router) insertAndReturnRoutingEntry(mainIPFlowKey shila.IPFlowKey, dstAddr shila.NetworkAddress) Entry {

	// Create new entry and insert it into the routing table
	newEntry := Entry{ Dst: dstAddr, Paths: nil }
	router.entries[mainIPFlowKey] = &newEntry
	
	return newEntry
}

func (router *Router) InsertDestinationFromIPAddressPortKey(key shila.IPAddressPortKey, dstAddr shila.NetworkAddress) error {
	if _, ok := router.fixedTable[key]; ok {
		return shila.TolerableError("Entry already exists.")
	} else {
		router.fixedTable[key] = dstAddr
		return nil
	}
}

func (router *Router) InsertEndpointTokenToIPFlow(p *shila.Packet) error {
	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return shila.PrependError(err, fmt.Sprint("Unable to convert token from key."))
			} else {
				if _, ok := router.mainIPFlows[token]; ok {
					return shila.TolerableError("Response already exists.")
				} else {
					router.mainIPFlows[token] = p.Flow.IPFlow.Key()
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

func (router *Router) getMainIPFlowKeyFromEndpointToken(packet *shila.Packet) (shila.IPFlowKey, bool) {
	// If the packet contains a receiver token, then the new connection is a sub flow.
	if token, err := mptcp.GetReceiverToken(packet.Payload); err == nil {
		mainIPFlowKey, ok := router.mainIPFlows[token]
		return mainIPFlowKey, ok
	}
	return "", false
}

func (router *Router) routeSubFlow(key shila.IPFlowKey) (Response, error) {

	if entry, ok := router.entries[key]; ok {
		return Response{
			Dst: 			entry.Dst,
			FlowCategory: 	SubFlow,
			MainIPFlowKey:  key,
		}, nil
	}

	return Response{}, GeneralError("Unable to route sub flow.")
}

func (router *Router) getDestinationFromIPOptions(packet *shila.Packet) (shila.NetworkAddress, bool) {
	return nil, false
}

func (router *Router) getDestinationFromIPAddressPortKey(packet *shila.Packet) (shila.NetworkAddress, bool) {
	dstAddr, ok := router.fixedTable[packet.Flow.IPFlow.DstKey()]
	return dstAddr, ok
}