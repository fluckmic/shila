package shila

// Should be able to create the TCPAddress
// from an arbitrary number of strings
type NetworkAddress interface {
	New(...string) error
	String() string
}

// Should be able to create the path
// from an arbitrary number of strings
type NetworkPath interface {
	New(...string) error
	String() string
}

type ClientNetworkEndpoint interface {
	New(connectTo NetworkAddress, connectVia NetworkPath, ingressBuf PacketChannel, egressBuf PacketChannel) ClientNetworkEndpoint
	Endpoint
	// TODO: config?
	// TODO: path renegotiation?
}

type ServerNetworkEndpoint interface {
	New(listenTo NetworkAddress, ingressBuf PacketChannel, egressBuf PacketChannel) ServerNetworkEndpoint
	Endpoint
}