package networkEndpoint

import (
	"net"
	"shila/core/shila"
)

type controlMessage struct {
	TcpFlow         shila.TCPFlow
	LAddrContactEnd net.UDPAddr
	Payload         []byte
}

type payloadMessage struct {
	Payload []byte
}