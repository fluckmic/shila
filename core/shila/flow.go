//
package shila

import (
	"fmt"
	"net"
	"shila/layer/tcpip"
	"strings"
)

type Flow struct {
	IPFlow  IPFlow
	NetFlow NetFlow
	Kind 	FlowKind
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

type FlowKind uint8
const (
	_                 = iota
	Mainflow FlowKind = iota
	Subflow
	Unknown
)

func(fk FlowKind) String() string {
	switch fk {
	case Mainflow: 	return "MainFlow"
	case Subflow:  	return "SubFlow"
	case Unknown:   return "Unknown"
	}
	return "Unknown"
}

func GetIPFlow(raw []byte) (IPFlow, error) {
	if src, dst, err := tcpip.DecodeSrcAndDstTCPAddr(raw); err != nil {
		return IPFlow{}, err
	} else {
		return IPFlow{Src: src, Dst: dst}, nil
	}
}

func GetIPFlowFromString(s string) (IPFlow, error) {

	flow := IPFlow{}

	if 	!strings.HasPrefix(s, KeyPrefix) {
		return flow, TolerableError(fmt.Sprint("Flow string has to start with {", KeyPrefix, "}."))
	} else {
		s = strings.TrimPrefix(s, KeyPrefix)
	}
	if 	!strings.HasSuffix(s, KeySuffix) {
		return flow, TolerableError(fmt.Sprint("Flow string has to end with {", KeySuffix, "}."))
	} else {
		s = strings.TrimSuffix(s, KeySuffix)
	}

	split := strings.Split(s, KeyDelimiter)

	if len(split) != 2 {
		return flow, TolerableError(fmt.Sprint("Flow string has to contain a delimiter {", KeyDelimiter, "}."))
	}

	src, err := tcpip.DecodeTCPAddrFromString(split[0])
	if err != nil {
		return flow, PrependError(err, "Cannot parse src address.")
	}
	dst, err := tcpip.DecodeTCPAddrFromString(split[1])
	if err != nil {
		return flow, PrependError(err, "Cannot parse dst address.")
	}

	flow.Dst = dst
	flow.Src = src

	return flow, nil

}