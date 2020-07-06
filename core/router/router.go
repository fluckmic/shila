package router

import (
	"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
	"shila/log"
	"sync"
)

type Router struct {
	mainIPFlows 	map[mptcp.EndpointToken] shila.IPFlow	    		// endpoint token to main ip flow keys
	endpointToken 	map[shila.IPFlowKey] mptcp.EndpointToken			// maps main ip flows to endpoint token
	entries     	map[shila.IPFlowKey] *Entry     					// ip flow keys to routing entries
	fixedTable 	 	map[shila.IPAddressPortKey] shila.NetworkAddress 	// ip address port key to destination address
	lock        	sync.Mutex
}


func New() Router {

	router := Router{
		mainIPFlows:	make(map[mptcp.EndpointToken] shila.IPFlow),
		endpointToken:	make(map[shila.IPFlowKey] mptcp.EndpointToken),
		entries: 		make(map[shila.IPFlowKey] *Entry),
		fixedTable:		make(map[shila.IPAddressPortKey] shila.NetworkAddress),
	}

	// See whether there is some routing from it which can be loaded
	if err := router.fillWithEntriesFromDisk(); err != nil {
		log.Error.Print(err.Error())
	}
	return router
}

func (router *Router) Route(packet *shila.Packet) (Response, error) {

	router.lock.Lock()
	defer router.lock.Unlock()

	if mainIPFlow, ok := router.getMainIPFlowFromEndpointToken(packet); ok {
		return router.routeSubFlow(packet, mainIPFlow)
	}

	return router.routeMainFlow(packet)
}

func (router *Router) InsertDestinationFromIPAddressPortKey(key shila.IPAddressPortKey, dstAddr shila.NetworkAddress) error {

	router.lock.Lock()
	defer router.lock.Unlock()

	if _, ok := router.fixedTable[key]; ok {
		return shila.TolerableError("Entry already exists.")
	} else {
		router.fixedTable[key] = dstAddr
		return nil
	}
}

func (router *Router) InsertEndpointTokenToIPFlow(p *shila.Packet) error {

	router.lock.Lock()
	defer router.lock.Unlock()

	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return shila.PrependError(err, fmt.Sprint("Unable to convert token from key."))
			} else {
				if _, ok := router.mainIPFlows[token]; ok {
					return shila.TolerableError("Response already exists.")
				} else {
					ipFLowKey := p.Flow.IPFlow.Key()
					router.mainIPFlows[token] 		= p.Flow.IPFlow
					router.endpointToken[ipFLowKey] = token
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

func (router *Router) ClearEntry(key shila.IPFlowKey) {

	router.lock.Lock()
	defer router.lock.Unlock()

	// First fetch the routing entry (if there is one), free the path and remove it from the path
	if entry, ok := router.entries[key]; ok {
		entry.Paths.free(key)
	}
	delete(router.entries, key)

	// For main flows remove the temporary entries
	if token, ok := router.endpointToken[key]; ok {
		delete(router.mainIPFlows, token)
	}
	delete(router.endpointToken, key)

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
		return PrependError(err, "Unable to load routing entries from disk.")
	}
	err = router.batchInsert(routingEntries)
	return nil
}

func (router *Router) getMainIPFlowFromEndpointToken(packet *shila.Packet) (shila.IPFlow, bool) {
	// If the packet contains a receiver token, then the new connection is a sub flow.
	if token, err := mptcp.GetReceiverToken(packet.Payload); err == nil {
		mainIPFlow, ok := router.mainIPFlows[token]
		return mainIPFlow, ok
	}
	return shila.IPFlow{}, false
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
		pathWrapper, flowCount := entry.Paths.get(mainIPFlowKey)
		return Response{
			Dst:          entry.Dst,
			FlowCategory: MainFlow,
			MainIPFlow:   packet.Flow.IPFlow,
			FlowCount:    flowCount,
			Path:         pathWrapper.path,
			RawMetrics:   pathWrapper.rawMetrics,
		},nil
	}

	return Response{}, shila.TolerableError("Unable to route packet. No routing information available.")
}

func (router *Router) routeSubFlow(packet *shila.Packet, ipFlow shila.IPFlow) (Response, error) {

	if entry, ok := router.entries[ipFlow.Key()]; ok {

		// Add add a link to the entry
		router.entries[packet.Flow.IPFlow.Key()] = entry

		// Create and return the response
		pathWrapper, subFlowCount := entry.Paths.get(packet.Flow.IPFlow.Key())
		return Response{
			Dst:          entry.Dst,
			FlowCategory: SubFlow,
			MainIPFlow:   ipFlow,
			FlowCount:    subFlowCount,
			Path:         pathWrapper.path,
			RawMetrics:   pathWrapper.rawMetrics,
		}, nil
	}

	return Response{}, GeneralError("Unable to route sub flow.")
}

func (router *Router) insertAndReturnRoutingEntry(mainIPFlowKey shila.IPFlowKey, dstAddr shila.NetworkAddress) Entry {

	// Create new entry and insert it into the routing table
	newEntry := Entry{ Dst: dstAddr, Paths: newPaths(dstAddr)}
	router.entries[mainIPFlowKey] = &newEntry
	return newEntry
}

func (router *Router) getDestinationFromIPOptions(packet *shila.Packet) (shila.NetworkAddress, bool) {
	return nil, false
}

func (router *Router) getDestinationFromIPAddressPortKey(packet *shila.Packet) (shila.NetworkAddress, bool) {
	dstAddr, ok := router.fixedTable[packet.Flow.IPFlow.DstKey()]
	return dstAddr, ok
}

