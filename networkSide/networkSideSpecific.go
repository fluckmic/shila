package networkSide

import (
	"net"
	"shila/core/shila"
	"shila/networkSide/network"
	"shila/networkSide/networkEndpoint"
	"strconv"
)

var _ shila.SpecificNetworkSideManager = (*SpecificManager)(nil)

type SpecificManager struct { }

func NewSpecificManager() SpecificManager {
	return SpecificManager{	}
}

func (specMng SpecificManager) NewClient(flow shila.Flow, label shila.EndpointLabel, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkClientEndpoint {
	return networkEndpoint.NewClient(flow, label, endpointIssues)
}

func (specMng SpecificManager) NewServer(flow shila.Flow, label shila.EndpointLabel, endpointIssues shila.EndpointIssuePubChannel) shila.NetworkServerEndpoint {
	return networkEndpoint.NewServer(flow, label, endpointIssues)
}

func (specMng SpecificManager) RemoteContactingFlow(flow shila.NetFlow) shila.NetFlow {
	return shila.NetFlow{
		Src:  flow.Src,
		Path: specMng.getDefaultContactingPath(flow.Dst),
		Dst:  specMng.generateRemoteContactingAddress(flow.Dst),
	}
}

func (specMng SpecificManager) LocalContactingNetFlow() shila.NetFlow {
	src, _ := network.AddressGenerator{}.NewLocal(strconv.Itoa(Config.ContactingServerPort))
	return shila.NetFlow{
		Src: 	src,
		Path: 	network.PathGenerator{}.NewEmpty(),
		Dst: 	network.AddressGenerator{}.NewEmpty(),
	}
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
