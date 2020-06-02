package networkEndpoint

import (
	"net"
	"shila/core/shila"
)

type networkConnection struct {
	Label            shila.EndpointLabel
	TrueNetFlow      shila.NetFlow
	RepresentingFlow shila.Flow
	Backbone         *net.TCPConn
}

type controlMessage struct {
	IPFlow                 shila.IPFlow
	srcAddrContactEndpoint string
}