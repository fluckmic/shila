//
package shila

type PacketChannel 			 chan *Packet
type EndpointIssuePubChannel chan EndpointIssuePub
type PacketChannelPubChannel chan PacketChannelPub

type PacketChannelPub struct {
	Publisher Endpoint
	Channel   PacketChannel
}

type EndpointIssuePubChannels struct {
	Ingress EndpointIssuePubChannel
	Egress  EndpointIssuePubChannel
}


type PacketChannelPubChannels struct {
	Ingress PacketChannelPubChannel
	Egress 	PacketChannelPubChannel
}

type EndpointIssuePub struct {
	Issuer Endpoint
	Key    TCPFlowKey
	Error  error
}

type PacketChannels struct {
	Ingress PacketChannel
	Egress  PacketChannel
}