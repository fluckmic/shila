package networkSide

import (
	"github.com/scionproto/scion/go/lib/snet"
	"net"
	"shila/config"
	"shila/core/shila"
	"shila/networkSide/networkEndpoint"
)

var _ shila.SpecificNetworkSideManager = (*SpecificManager)(nil)

type SpecificManager struct { }

func NewSpecificManager() SpecificManager {
	return SpecificManager{	}
}

func (specMng SpecificManager) NewContactClient(rAddr shila.NetworkAddress, path shila.NetworkPath, tcpFlow shila.TCPFlow, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewContactClient(rAddr, path, tcpFlow, endpointIssues)
}

func (specMng SpecificManager) NewTrafficClient(lAddrContactEnd shila.NetworkAddress, rAddr shila.NetworkAddress, path shila.NetworkPath,
	tcpFlow shila.TCPFlow, issues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewTrafficClient(lAddrContactEnd, rAddr, path, tcpFlow, issues)
}

func (specMng SpecificManager) NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return networkEndpoint.NewServer(lAddr, role, issues)
}

func (specMng SpecificManager) ContactRemoteAddr(rAddressTraffic shila.NetworkAddress) shila.NetworkAddress {
	rAddressContact := rAddressTraffic.(*snet.UDPAddr).Copy()
	rAddressContact.Host.Port = config.Config.NetworkSide.ContactingServerPort
	return rAddressContact
}

func (specMng SpecificManager) ContactLocalAddr() shila.NetworkAddress {
	return &snet.UDPAddr{Host: &net.UDPAddr{Port: config.Config.NetworkSide.ContactingServerPort}}
}