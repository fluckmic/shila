//
package shila

const InitialTTL int = 5		// Number of sending retries for a packet

type Packet struct {
	Entrypoint 	Endpoint
	Flow       	Flow
	TTL			int
	Payload    	[]byte
}

func NewPacket(ep Endpoint, ipf IPFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{IPFlow: ipf}, TTL: InitialTTL, Payload: raw}
}

func NewPacketWithNetFlowAndKind(ep Endpoint, ipf IPFlow, nf NetFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{IPFlow: ipf, NetFlow: nf}, TTL: InitialTTL, Payload: raw}
}