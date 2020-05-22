package shila

// Mappings
type ServerNetworkEndpointMapping struct {
	ServerNetworkEndpoint
	IPConnectionMapping
}

type IPConnectionMapping 		map[IPFlowKey]	bool
type ServerEndpointMapping		map[NetworkAddressKey]			ServerNetworkEndpointMapping
type ClientEndpointMapping 		map[IPFlowKey]					ClientNetworkEndpoint

func (sb *ServerNetworkEndpointMapping) AddIPFlowKey(key IPFlowKey) {
	sb.IPConnectionMapping[key] = true
}

func (sb *ServerNetworkEndpointMapping) RemoveIPFlowKey(key IPFlowKey) {
	delete(sb.IPConnectionMapping, key)
}

func (sb *ServerNetworkEndpointMapping) Empty() bool {
	return len(sb.IPConnectionMapping) == 0
}
