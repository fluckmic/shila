package shila

import (
	"net"
)

type Flow struct {
	IPFlow  IPFlow
	NetFlow NetFlow
}

// Has to be parsed for every packet
type IPFlow struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

type NetFlow struct {
	Src  NetworkAddress
	Path NetworkPath
	Dst  NetworkAddress
}


