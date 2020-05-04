package shila

type PacketChannel chan *Packet

type TrafficChannels struct {
	Ingress PacketChannel
	Egress  PacketChannel
}

type ContactingChannel struct {
	Endpoint ClientNetworkEndpoint
	Channels TrafficChannels
}
