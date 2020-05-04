package shila

type EndpointLabel uint8

const (
	_                 		    	 		  = iota
	KernelEndpoint 				EndpointLabel = iota
	ContactingNetworkEndpoint
	TrafficNetworkEndpoint
)

type Endpoint interface {
	TearDown() 			error
	Label() 			EndpointLabel
	TrafficChannels()	TrafficChannels
}