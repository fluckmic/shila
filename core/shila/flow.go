//
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

func (ipf IPFlow) Swap() IPFlow {
	return IPFlow{
		Src: ipf.Dst,
		Dst: ipf.Src,
	}
}

type NetFlow struct {
	Src   NetworkAddress
	Path  NetworkPath
	Dst   NetworkAddress
}

func (nf NetFlow) Swap() NetFlow {
	return NetFlow{
		Src:  nf.Dst,
		Path: nf.Path,
		Dst:  nf.Src,
	}
}

func GetIPFlow(raw []byte) (IPFlow, error) {
	if src, dst, err := tcpip.DecodeSrcAndDstTCPAddr(raw); err != nil {
		return IPFlow{}, err
	} else {
		return IPFlow{Src: src, Dst: dst}, nil
	}
}

