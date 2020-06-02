package networkEndpoint

import (
	"net"
	"shila/core/shila"
)

type networkConnection struct {
	Label            shila.EndpointLabel
	TrueNetFlow      shila.NetFlow				// The network flow corresponding to the backbone connection.
	RepresentingFlow shila.Flow					// The network and ip flow represented by this backbone connection.
	Backbone         *net.TCPConn
}

type controlMessage struct {
	IPFlow                 shila.IPFlow
	SrcAddrContactEndpoint net.TCPAddr
}