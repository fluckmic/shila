package model

type PacketChannel chan *Packet

type TrafficChannels struct {
	Label   EndpointLabel
	Key     EndpointKey
	Ingress PacketChannel
	Egress  PacketChannel
}

