package shila

type EndpointLabel uint8

const (
	_                 		     = iota
	KernelEndpoint EndpointLabel = iota
)

const ()

type endpoint interface {
	Setup() 	error
	TearDown() 	error
	Label() 	EndpointLabel
}

type clientNetworkEndpoint interface {
	New(connectTo address, connectVia path, ingressBuf chan *packet, egressBuf chan *packet) clientNetworkEndpoint
	endpoint
	// TODO: config?
	// TODO: path renegotiation?
}

type serverNetworkEndpoint interface {
	New(listenTo address, ingressBuf chan *packet, egressBuf chan *packet) serverNetworkEndpoint
	endpoint
}

type packet interface {
	SetAddress(address address)
	GetAddress() address

	SetPayload(payload IPv4TCPPacket)
	GetPayload() IPv4TCPPacket
}

// Should be able to create the address
// from an arbitrary number of strings
type address interface {
	New(...string) error
	String() string
}

// Should be able to create the path
// from an arbitrary number of strings
type path interface {
	New(...string) error
	String() string
}
