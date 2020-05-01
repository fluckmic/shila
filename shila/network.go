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

type NetworkEndpoint interface {
	NewClient(connectTo NetworkAddress, connectVia NetworkPath, ingressBuf PacketChannel, egressBuf PacketChannel) ClientNetworkEndpoint
	NewServer(listenTo NetworkAddress, ingressBuf PacketChannel, egressBuf PacketChannel) ServerNetworkEndpoint
}

type ClientNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
}

type ServerNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
}