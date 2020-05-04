package shila

type EndpointLabel uint8

const (
	_                 		     		  = iota
	KernelEndpoint 			EndpointLabel = iota
	NetworkClientEndpoint
	NetworkServerEndpoint
)

type Endpoint interface {
	TearDown() 	error
	Label() 	EndpointLabel
}