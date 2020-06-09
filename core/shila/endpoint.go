//
package shila

type EndpointIdentifier string
type EndpointRole uint8

const (
	_                                  = iota
	IngressKernelEndpoint EndpointRole = iota
	EgressKernelEndpoint
	ContactingNetworkEndpoint
	TrafficNetworkEndpoint
)

func (el EndpointRole) String() string {
	switch el {
		case ContactingNetworkEndpoint: return "ContactingNetworkEndpoint"
		case TrafficNetworkEndpoint:	return "TrafficNetworkEndpoint"
		case IngressKernelEndpoint: 	return "IngressKernelEndpoint"
		case EgressKernelEndpoint: 		return "EgressKernelEndpoint"
	}
	return "Unknown"
}

// Interface which each endpoint (kernel and network side) should implement.
type Endpoint interface {
	TearDown() error
	Role() 				EndpointRole
	TrafficChannels() 	PacketChannels
	Identifier() 		EndpointIdentifier
	Says(string)		string
}
