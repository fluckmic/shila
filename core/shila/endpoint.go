package shila

type EndpointLabel uint8

const (
	_                            = iota
	KernelEndpoint EndpointLabel = iota
	ContactingNetworkEndpoint
	TrafficNetworkEndpoint
)

func (el EndpointLabel) String() string {
	switch el {
		case KernelEndpoint: 			return "KernelEndpoint"
		case ContactingNetworkEndpoint: return "ContactingNetworkEndpoint"
		case TrafficNetworkEndpoint:	return "TrafficNetworkEndpoint"
		default:						return "Unknown"
	}
}

type Endpoint interface {
	TearDown() error
	Label() 			EndpointLabel
	Key() 				EndpointKey
	TrafficChannels() 	PacketChannels
}

