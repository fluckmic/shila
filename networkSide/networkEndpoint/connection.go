package networkEndpoint

import (
	"net"
	"shila/core/shila"
)

type networkConnection struct {
	EndpointRole     shila.EndpointRole
	TrueNetFlow      shila.NetFlow				// The network lAddress corresponding to the backbone connection.
	RepresentingFlow shila.Flow					// The network and ip lAddress represented by this backbone connection.
	Backbone         *net.TCPConn				// FIXME: *snet.Conn
}

type controlMessage struct {
	IPFlow                 shila.IPFlow
	SrcAddrContactEndpoint net.TCPAddr
}