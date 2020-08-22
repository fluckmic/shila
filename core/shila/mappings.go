//
package shila

// Mappings
type MappingNetworkServerEndpoint map[NetworkAddressKey]NetworkServerEndpointTCPFlowRegister
type MappingNetworkClientEndpoint map[TCPFlowKey] NetworkClientEndpoint

// Register of TCP flow keys
type TCPFlowRegister map[TCPFlowKey]	bool

// Contains all registered tcp flows for a given network server endpoint
type NetworkServerEndpointTCPFlowRegister struct {
	NetworkServerEndpoint
	TCPFlowRegister
}

func NewNetworkServerEndpointTCPFlowRegister(endpoint NetworkServerEndpoint) NetworkServerEndpointTCPFlowRegister {
	return NetworkServerEndpointTCPFlowRegister{
		NetworkServerEndpoint: endpoint,
		TCPFlowRegister:       make(TCPFlowRegister),
	}
}

func (r *NetworkServerEndpointTCPFlowRegister) Register(key TCPFlowKey) {
	r.TCPFlowRegister[key] = true
}

func (r *NetworkServerEndpointTCPFlowRegister) Unregister(key TCPFlowKey) {
	delete(r.TCPFlowRegister, key)
}

func (r *NetworkServerEndpointTCPFlowRegister) IsEmpty() bool {
	return len(r.TCPFlowRegister) == 0
}
