//
package shila

import (
	"net"
	"shila/layer/tcpip"
)

type Flow struct {
	TCPFlow TCPFlow
	NetFlow NetFlow
}

// Has to be parsed for every packet
type TCPFlow struct {
	Src net.TCPAddr
	Dst net.TCPAddr
}

func (ipf TCPFlow) Swap() TCPFlow {
	return TCPFlow{
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

func GetTCPFlow(raw []byte) (TCPFlow, error) {
	if src, dst, err := tcpip.DecodeSrcAndDstTCPAddr(raw); err != nil {
		return TCPFlow{}, err
	} else {
		return TCPFlow{Src: src, Dst: dst}, nil
	}
}

