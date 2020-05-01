package shila

type EndpointLabel uint8

const (
	_                 		     		  = iota
	KernelEndpoint 			EndpointLabel = iota
	NetworkClientEndpoint
	NetworkServerEndpoint
)

type Endpoint interface {
	Setup() 	error
	TearDown() 	error
	Label() 	EndpointLabel
}