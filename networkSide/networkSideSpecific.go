package networkSide

import (
	"github.com/scionproto/scion/go/lib/snet"
	"net"
	"shila/core/shila"
	"shila/networkSide/network"
	"shila/networkSide/networkEndpoint"
)

var _ shila.SpecificNetworkSideManager = (*SpecificManager)(nil)

type SpecificManager struct { }

func NewSpecificManager() SpecificManager {
	return SpecificManager{	}
}

func (specMng SpecificManager) NewClient(flow shila.Flow, label shila.EndpointRole, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewClient(flow, label, endpointIssues)
}

func (specMng SpecificManager) NewServer(lAddr shila.NetworkAddress, role shila.EndpointRole, issues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return networkEndpoint.NewServer(lAddr, role, issues)
}

func (specMng SpecificManager) ContactRemoteAddr(flow shila.NetFlow) shila.NetFlow {
	return shila.NetFlow{
		Src:  flow.Src,
		Path: specMng.getDefaultContactingPath(flow.Dst),
		Dst:  specMng.generateRemoteContactingAddress(flow.Dst),
	}
}

func (specMng SpecificManager) ContactLocalAddr() shila.NetworkAddress {
	return &snet.UDPAddr{Host: &net.UDPAddr{Port: Config.ContactingServerPort}}
}

func (specMng SpecificManager) generateRemoteContactingAddress(address shila.NetworkAddress) shila.NetworkAddress {
	addr := address.(*net.TCPAddr)
	return &net.TCPAddr{
		IP:   addr.IP,
		Port: Config.ContactingServerPort,
		Zone: addr.Zone,
	}
}

func (specMng SpecificManager) getDefaultContactingPath(address shila.NetworkAddress) shila.NetworkPath {
	_ = address
	return network.PathGenerator{}.NewEmpty()
}
