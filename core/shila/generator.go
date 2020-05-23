package shila

// Defines all the interfaces which the network endpoint generator has to
// implement as they are used by the manager of the network side.

type NetworkEndpointGenerator interface {
	NewClient(netConnId NetFlow, l EndpointLabel) ClientNetworkEndpoint
	NewServer(netConnId NetFlow, l EndpointLabel) ServerNetworkEndpoint
}

type NetworkNetFlowGenerator interface {
	LocalContactingNetFlow() 		NetFlow
	RemoteContactingFlow(NetFlow) 	NetFlow
}

type NetworkAddressGenerator interface {
	New(string)		 NetworkAddress 	// Generates a new network address from a string.
	NewLocal(string) NetworkAddress		// Generates a new local network address from a string.
	NewEmpty()		 NetworkAddress 	// Generates a empty network address.
}

type NetworkPathGenerator interface {
	New(string)		NetworkPath
}

type ClientNetworkEndpoint interface {
	Endpoint
	SetupAndRun() 	(NetFlow, error)
}

type ServerNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
}

type NetworkAddress interface {
	String() string
}

type NetworkPath interface {
	String() string
}
