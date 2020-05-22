package shila

type Packet struct {
	Entrypoint Endpoint
	Flow       Flow
	Payload    []byte
}

func NewPacket(ep Endpoint, ipf IPFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{IPFlow: ipf}, Payload: raw}
}

func NewPacketWithNetFlow(ep Endpoint, ipf IPFlow, nf NetFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{IPFlow: ipf, NetFlow: nf}, Payload: raw}
}