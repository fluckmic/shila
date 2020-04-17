package network

import "shila/shila"

type Error string

func (e Error) Error() string {
	return string(e)
}

type endpoint interface {
	Setup() error
	TearDown() error
}

type clientEndpoint interface {
	New(connectTo address, connectVia path, ingressBuf chan *packet, egressBuf chan *packet)
	endpoint
	// TODO: config?
	// TODO: path renegotiation?
}

type serverEndpoint interface {
	New(listenTo address, ingressBuf chan *packet, egressBuf chan *packet)
	endpoint
}

type packet interface {
	SetAddress(address address)
	GetAddress() address

	SetPayload(payload shila.IPPacket)
	GetPayload() shila.IPPacket
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
