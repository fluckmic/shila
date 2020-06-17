//
package router

import (
	"fmt"
	"shila/core/shila"
	"shila/layer/mptcp"
)



func (router *Router) getFromIPOptions(raw []byte) (shila.NetFlow, bool, error) {
	return shila.NetFlow{}, false, nil
}

func (router *Router) getFromIPAddressPortKey(key shila.IPAddressPortKey) (*shila.NetFlow, bool) {
	packetHeader, ok := router.destinations.fromIPPortKey[key]
	return packetHeader, ok
}

func (router *Router) getFromMPTCPEndpointToken(token mptcp.EndpointToken) (*shila.Flow, bool) {
	mainIPFlowKey, ok := router.destinations.fromMPTCPToken[token]

	return packetHeader, ok
}


}