//
package shila

type EndpointRole uint8

const (
	_                                  = iota
	IngressKernelEndpoint EndpointRole = iota
	EgressKernelEndpoint
	ContactNetworkEndpoint
	TrafficNetworkEndpoint
)

func (el EndpointRole) String() string {
	switch el {
		case ContactNetworkEndpoint: 	return "Contact Network Endpoint"
		case TrafficNetworkEndpoint:	return "Traffic Network Endpoint"
		case IngressKernelEndpoint: 	return "Ingress Kernel Endpoint"
		case EgressKernelEndpoint: 		return "Egress Kernel Endpoint"
	}
	return "Unknown"
}

// Interface which each endpoint (kernel and network side) should implement.
type Endpoint interface {
	TearDown() error
	Role() 				EndpointRole
	TrafficChannels() 	PacketChannels
	Identifier() 		string
	Says(string)		string
}
