//
package shila

type PacketChannel 			 chan *Packet
type EndpointIssuePubChannel chan EndpointIssuePub
type PacketChannelPubChannel chan PacketChannelPub

type PacketChannelPub struct {
	Publisher Endpoint
	Channel   PacketChannel
}

type EndpointIssuePub struct {
	Publisher Endpoint
	Flow	  Flow
	Error 	  error
}

type PacketChannels struct {
	Ingress PacketChannel
	Egress  PacketChannel
}