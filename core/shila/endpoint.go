//
package shila

type EndpointLabel uint8

const (
	_                            = iota
	IngressKernelEndpoint EndpointLabel = iota
	EgressKernelEndpoint
	ContactingNetworkEndpoint
	TrafficNetworkEndpoint
)

func (el EndpointLabel) String() string {
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
	Label() 			EndpointLabel
	Key() 				EndpointKey
	TrafficChannels() 	PacketChannels
}

