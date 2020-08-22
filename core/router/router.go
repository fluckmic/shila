package router

import (
	"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
	"shila/log"
	"sync"
)

type Router struct {
	mainTCPFlows  map[mptcp.EndpointToken] shila.TCPFlow           // endpoint token to main tcp flow keys
	endpointToken map[shila.TCPFlowKey] mptcp.EndpointToken        // maps main tcp flows to endpoint token
	entries       map[shila.TCPFlowKey] *Entry                     // tcp flow keys to routing entries
	fixedTable    map[shila.IPAddressPortKey] shila.NetworkAddress // ip address port key to destination address
	lock          sync.Mutex
}


func New() Router {

	router := Router{
		mainTCPFlows:  make(map[mptcp.EndpointToken] shila.TCPFlow),
		endpointToken: make(map[shila.TCPFlowKey] mptcp.EndpointToken),
		entries:       make(map[shila.TCPFlowKey] *Entry),
		fixedTable:    make(map[shila.IPAddressPortKey] shila.NetworkAddress),
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

	if mainTCPFlow, ok := router.getMainTCPFlowFromEndpointToken(packet); ok {
		return router.routeSubFlow(packet, mainTCPFlow)
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

func (router *Router) InsertEndpointTokenToTCPFlow(p *shila.Packet) error {

	router.lock.Lock()
	defer router.lock.Unlock()

	if key, ok, err := mptcp.GetSenderKey(p.Payload); ok {
		if err == nil {
			if token, err := mptcp.EndpointKeyToToken(key); err != nil {
				return shila.PrependError(err, fmt.Sprint("Unable to convert token from key."))
			} else {
				if _, ok := router.mainTCPFlows[token]; ok {
					return shila.TolerableError("Response already exists.")
				} else {
					tcpFlowKey := p.Flow.TCPFlow.Key()
					router.mainTCPFlows[token] 		= p.Flow.TCPFlow
					router.endpointToken[tcpFlowKey] = token
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

func (router *Router) ClearEntry(key shila.TCPFlowKey) {

	router.lock.Lock()
	defer router.lock.Unlock()

	// First fetch the routing entry (if there is one), free the path and remove it from the path
	if entry, ok := router.entries[key]; ok {
		entry.Paths.free(key)
	}
	delete(router.entries, key)

	// For main flows remove the temporary entries
	if token, ok := router.endpointToken[key]; ok {
		delete(router.mainTCPFlows, token)
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

func (router *Router) getMainTCPFlowFromEndpointToken(packet *shila.Packet) (shila.TCPFlow, bool) {
	// If the packet contains a receiver token, then the new connection is a sub flow.
	if token, err := mptcp.GetReceiverToken(packet.Payload); err == nil {
		mainTCPFlow, ok := router.mainTCPFlows[token]
		return mainTCPFlow, ok
	}
	return shila.TCPFlow{}, false
}

func (router *Router) routeMainFlow(packet *shila.Packet) (Response, error) {

	// The key we get directly from the packet
	mainTCPFlowKey := packet.Flow.TCPFlow.Key()

	// For the destination we have to options:
	var dstAddr shila.NetworkAddress; var ok bool

	// 1) From the IP Options
	dstAddr, ok = router.getDestinationFromIPOptions(packet)
	if !ok {
		// 2) From the routing table
		dstAddr, ok = router.getDestinationFromIPAddressPortKey(packet)
	}

	if ok {

		if entry, err := router.insertAndReturnRoutingEntry(mainTCPFlowKey, dstAddr); err != nil {
			return Response{}, shila.PrependError(err, "Unable to route packet.")

		} else {

			pathWrapper, flowCount := entry.Paths.get(mainTCPFlowKey)

			return Response{
				Dst:          entry.Dst,
				FlowCategory: MainFlow,
				MainTCPFlow:  packet.Flow.TCPFlow,
				FlowCount:    flowCount,
				Path:         pathWrapper.path,
				RawMetrics:   pathWrapper.rawMetrics,
				Sharability:  entry.Paths.sharability,
			},nil
		}
	}

	return Response{}, shila.TolerableError("Unable to route packet. No routing information available.")
}

func (router *Router) routeSubFlow(packet *shila.Packet, tcpFlow shila.TCPFlow) (Response, error) {

	if entry, ok := router.entries[tcpFlow.Key()]; ok {

		// Add add a link to the entry
		router.entries[packet.Flow.TCPFlow.Key()] = entry

		// Create and return the response
		pathWrapper, subFlowCount := entry.Paths.get(packet.Flow.TCPFlow.Key())
		return Response{
			Dst:          entry.Dst,
			FlowCategory: SubFlow,
			MainTCPFlow:  tcpFlow,
			FlowCount:    subFlowCount,
			Path:         pathWrapper.path,
			RawMetrics:   pathWrapper.rawMetrics,
			Sharability:  entry.Paths.sharability,
		}, nil
	}

	return Response{}, GeneralError("Unable to route sub flow.")
}

func (router *Router) insertAndReturnRoutingEntry(mainTCPFlowKey shila.TCPFlowKey, dstAddr shila.NetworkAddress) (Entry, error) {
	// Create new entry and insert it into the routing table
	if paths, err := newPaths(dstAddr); err != nil {
		return Entry{}, err
	} else {
		newEntry := Entry{ Dst: dstAddr, Paths: paths}
		router.entries[mainTCPFlowKey] = &newEntry
		return newEntry, nil
	}
}

func (router *Router) getDestinationFromIPOptions(packet *shila.Packet) (shila.NetworkAddress, bool) {
	return nil, false
}

func (router *Router) getDestinationFromIPAddressPortKey(packet *shila.Packet) (shila.NetworkAddress, bool) {
	dstAddr, ok := router.fixedTable[packet.Flow.TCPFlow.DstKey()]
	return dstAddr, ok
}

