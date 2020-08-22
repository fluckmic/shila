//
package shila

const InitialTTL int = 5		// Number of sending retries for a packet

type Packet struct {
	Entrypoint 	Endpoint
	Flow       	Flow
	TTL			int
	Payload    	[]byte
}

func NewPacket(ep Endpoint, ipf TCPFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{TCPFlow: ipf}, TTL: InitialTTL, Payload: raw}
}

func NewPacketWithNetFlowAndKind(ep Endpoint, ipf TCPFlow, nf NetFlow, raw []byte) *Packet {
	return &Packet{Entrypoint: ep, Flow: Flow{TCPFlow: ipf, NetFlow: nf}, TTL: InitialTTL, Payload: raw}
}