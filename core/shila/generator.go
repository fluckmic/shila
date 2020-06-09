//
package shila

// Defines all the interfaces which the network endpoint generator has to
// implement as they are used by the manager of the network side.

type SpecificNetworkSideManager interface {
	NewClient(flow Flow, r EndpointRole, c EndpointIssuePubChannel) NetworkClientEndpoint
	NewServer(flow Flow, r EndpointRole, c EndpointIssuePubChannel) NetworkServerEndpoint
	ContactLocalAddr() 								NetFlow
	ContactRemoteAddr(NetFlow) 					NetFlow
}

type NetworkAddressGenerator interface {
	New(string)		 (NetworkAddress, error) 	// Generates a new network address from a string.
	NewLocal(string) (NetworkAddress, error)	// Generates a new local network address from a string.
	NewEmpty()		 NetworkAddress 			// Generates a empty network address.
}

type NetworkPathGenerator interface {
	New(string)		(NetworkPath, error)
}

type NetworkClientEndpoint interface {
	Endpoint
	SetupAndRun() 	(NetFlow, error)
	Flow()			Flow
}

type NetworkServerEndpoint interface {
	Endpoint
	SetupAndRun() 	error
}

type NetworkAddress interface {
	String() string
}

type NetworkPath interface {
	String() string
}
