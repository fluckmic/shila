//
package shila

type Packet struct {
	Entrypoint 	Endpoint
	Flow       	Flow
	TTL			int
	Payload    	[]byte
}

func NewPacket(ep Endpoint, ipf IPFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{IPFlow: ipf}, TTL: Config.InitialTTL, Payload: raw}
}

func NewPacketWithNetFlow(ep Endpoint, ipf IPFlow, nf NetFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{IPFlow: ipf, NetFlow: nf}, TTL: Config.InitialTTL, Payload: raw}
}