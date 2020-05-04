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
	NewClient(connectTo NetworkAddress, connectVia NetworkPath, label EndpointLabel, channels TrafficChannels) ClientNetworkEndpoint
	NewServer(listenTo NetworkAddress, label EndpointLabel, channels TrafficChannels) ServerNetworkEndpoint
}

type ClientNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
}

type ServerNetworkEndpoint interface {
	Endpoint
	SetupAndRun() error
}