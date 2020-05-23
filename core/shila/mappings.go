package shila

// Mappings
type MappingNetworkServerEndpoint map[NetworkAddressKey] NetworkServerEndpointIPFlowRegister
type MappingNetworkClientEndpoint map[IPFlowKey] NetworkClientEndpoint

// Register of IP flow keys
type IPFlowRegister map[IPFlowKey]	bool

// Contains all registered IP flows for a given network server endpoint
type NetworkServerEndpointIPFlowRegister struct {
	NetworkServerEndpoint
	IPFlowRegister
}

func NewNetworkServerEndpointIPFlowRegister(endpoint NetworkServerEndpoint) NetworkServerEndpointIPFlowRegister {
	return NetworkServerEndpointIPFlowRegister{
		NetworkServerEndpoint: endpoint,
		IPFlowRegister:		   make(IPFlowRegister),
	}
}

func (r *NetworkServerEndpointIPFlowRegister) Register(key IPFlowKey) {
	r.IPFlowRegister[key] = true
}

func (r *NetworkServerEndpointIPFlowRegister) Unregister(key IPFlowKey) {
	delete(r.IPFlowRegister, key)
}

func (r *NetworkServerEndpointIPFlowRegister) IsEmpty() bool {
	return len(r.IPFlowRegister) == 0
}
