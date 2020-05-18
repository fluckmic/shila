package model

type PacketChannel chan *Packet

type PacketChannelAnnouncement struct {
	Announcer Endpoint
	Channel   PacketChannel
}

type PacketChannels struct {
	Ingress PacketChannel
	Egress  PacketChannel
}