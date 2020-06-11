package networkSide

import (
	"github.com/scionproto/scion/go/lib/snet"
	"net"
	"shila/core/shila"
	"shila/networkSide/networkEndpoint"
)

var _ shila.SpecificNetworkSideManager = (*SpecificManager)(nil)

type SpecificManager struct { }

func NewSpecificManager() SpecificManager {
	return SpecificManager{	}
}

func (specMng SpecificManager) NewContactClient(rAddr shila.NetworkAddress, ipFlow shila.IPFlow, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewContactClient(rAddr, ipFlow, endpointIssues)
}

func (specMng SpecificManager) NewTrafficClient(lAddrContactEnd shila.NetworkAddress, rAddr shila.NetworkAddress,
	ipFlow shila.IPFlow, issues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewTrafficClient(lAddrContactEnd, rAddr, ipFlow, issues)
}

func (specMng SpecificManager) NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return networkEndpoint.NewServer(lAddr, role, issues)
}

func (specMng SpecificManager) ContactRemoteAddr(rAddressTraffic shila.NetworkAddress) shila.NetworkAddress {
	rAddressContact := rAddressTraffic.(*snet.UDPAddr).Copy()
	rAddressContact.Host.Port = Config.ContactingServerPort
	return rAddressContact
}

func (specMng SpecificManager) ContactLocalAddr() shila.NetworkAddress {
	return &snet.UDPAddr{Host: &net.UDPAddr{Port: Config.ContactingServerPort}}
}