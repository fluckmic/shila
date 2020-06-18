//
package shila

// Defines all the interfaces which the network endpoint generator has to
// implement as they are used by the manager of the network side.

type SpecificNetworkSideManager interface {
	NewContactClient(rAddr NetworkAddress, ipFLow IPFlow, c EndpointIssuePubChannel) NetworkClientEndpoint
	NewTrafficClient(lAddrContactEnd NetworkAddress, rAddr NetworkAddress, ipFLow IPFlow, c EndpointIssuePubChannel) NetworkClientEndpoint
	NewServer(lAddr NetworkAddress, r EndpointRole, c EndpointIssuePubChannel) NetworkServerEndpoint
	ContactLocalAddr() 							NetworkAddress
	ContactRemoteAddr(NetworkAddress) 			NetworkAddress
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
}

type NetworkServerEndpoint interface {
	Endpoint
	SetupAndRun() 	error
}

type NetworkAddress interface {
	String() string
}

type NetworkPath interface {}
