//
package shila

import (
	"fmt"
	"net"
)

// Keys
type IPAddressKey 			  	string		// (ipv4)
type IPAddressPortKey		  	string		// (ipv4:port)
type IPFlowKey 					string		// (ipv4-address:port<>ipv4-address:port)

type NetworkAddressKey		  	string		// (network-address)
type NetworkAddressAndPathKey 	string		// (network-address<>path)
type NetFlowKey 				string		// (network-address<>path<>network-address)

type FlowKey					string		// (ip-flow-key,network-flow-key,flow-kind)

type EndpointKey   				string		//

type PacketKey					string 		// (endpoint-key,flow-key)

// Key generator
func GetIPAddressKey(ip net.IP) IPAddressKey {
	return IPAddressKey(fmt.Sprint("(",ip.String(),")"))
}

func GetIPAddressPortKey(addr net.TCPAddr) IPAddressPortKey {
	return IPAddressPortKey(fmt.Sprint("(", addr.String(), ")"))
}

func GetNetworkAddressAndPathKey(addr NetworkAddress, path NetworkPath) NetworkAddressAndPathKey {
	return NetworkAddressAndPathKey(fmt.Sprint("(", addr.String(),"<>",path.String(),")"))
}

func GetNetworkAddressKey(addr NetworkAddress) NetworkAddressKey {
	return NetworkAddressKey(fmt.Sprint("(", addr.String(), ")"))
}

func (ipf *IPFlow) Key() IPFlowKey {
	srcString := ipf.Src.String(); dstString := ipf.Dst.String()
	if srcString < dstString {
		return IPFlowKey(fmt.Sprint("(", srcString, "<>", dstString, ")"))
	} else {
		return IPFlowKey(fmt.Sprint("(", dstString, "<>", srcString, ")"))
	}
}

func (nf *NetFlow) Key() NetFlowKey {
	srcString := nf.Src.String(); dstString := nf.Dst.String()
	if srcString < dstString {
		return NetFlowKey(fmt.Sprint("(", srcString, "<>", nf.Path.String(), "<>", dstString, ")"))
	} else {
		return NetFlowKey(fmt.Sprint("(", dstString, "<>", nf.Path.String(), "<>", srcString, ")"))
	}
}

func (fl *Flow) Key() FlowKey {
	return FlowKey(fmt.Sprint("(", fl.IPFlow.Key(), ",", fl.NetFlow.Key(), ",", fl.Kind, ")"))
}

func (ipf *IPFlow) SrcKey() IPAddressPortKey {
	return GetIPAddressPortKey(ipf.Src)
}

func (ipf *IPFlow) SrcIPKey() IPAddressKey {
	return GetIPAddressKey(ipf.Src.IP)
}

func (ipf *IPFlow) DstKey() IPAddressPortKey {
	return GetIPAddressPortKey(ipf.Dst)
}

func (ipf *IPFlow) DstIPKey() IPAddressKey {
	return GetIPAddressKey(ipf.Dst.IP)
}

func (nf *NetFlow) DstAndPathKey() NetworkAddressAndPathKey {
	return GetNetworkAddressAndPathKey(nf.Dst, nf.Path)
}

func (p *Packet) Key() PacketKey {
	return PacketKey(fmt.Sprint("(", p.Entrypoint.Key(), ",", p.Flow.Key(), ")"))
}