package shila

import (
	"net"
	"shila/layer/tcpip"
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

func GetIPFlow(raw []byte) (IPFlow, error) {
	if src, dst, err := tcpip.DecodeSrcAndDstTCPAddr(raw); err != nil {
		return IPFlow{}, err
	} else {
		return IPFlow{Src: src, Dst: dst}, nil
	}
}

func GetNetFlowFromIPOptions(raw []byte) (NetFlow, bool, error) {
	return NetFlow{}, false, nil
}