//
package shila

import (
	"fmt"
	"net"
)

const (
	KeyPrefix    = "("
	KeyDelimiter = "|"
	KeySuffix    = ")"
)

// Keys
type IPAddressKey 			  	string		// (ipv4)
type IPAddressPortKey		  	string		// (ipv4:port)
type IPFlowKey 					string		// (ipv4-address:port<>ipv4-address:port)

type NetworkAddressKey		  	string		// (network-address)
type NetworkAddressAndPathKey 	string		// (network-address<>path)
type NetFlowKey 				string		// (network-address<>path<>network-address)

type FlowKey					string		// (ip-flow-key<>network-flow-key<>flow-kind)

type EndpointKey   				string		//

type PacketKey					string 		// (endpoint-key<>flow-key)

// Key generator
func GetIPAddressKey(ip net.IP) IPAddressKey {
	return IPAddressKey(fmt.Sprint(KeyPrefix,ip.String(), KeySuffix))
}

func GetIPAddressPortKey(addr net.TCPAddr) IPAddressPortKey {
	return IPAddressPortKey(fmt.Sprint(KeyPrefix, addr.String(), KeySuffix))
}

func GetNetworkAddressAndPathKey(addr NetworkAddress, path NetworkPath) NetworkAddressAndPathKey {
	return NetworkAddressAndPathKey(fmt.Sprint(KeyPrefix, addr.String(), KeyDelimiter,path.String(), KeySuffix))
}

func GetNetworkAddressKey(addr NetworkAddress) NetworkAddressKey {
	return NetworkAddressKey(fmt.Sprint(KeyPrefix, addr.String(), KeySuffix))
}

func (ipf *IPFlow) Key() IPFlowKey {
	srcString := ipf.Src.String(); dstString := ipf.Dst.String()
	if srcString < dstString {
		return IPFlowKey(fmt.Sprint(KeyPrefix, srcString, KeyDelimiter, dstString, KeySuffix))
	} else {
		return IPFlowKey(fmt.Sprint(KeyPrefix, dstString, KeyDelimiter, srcString, KeySuffix))
	}
}

func (ipf *IPFlow) String() string {
	srcString := ipf.Src.String(); dstString := ipf.Dst.String()
	return fmt.Sprint(KeyPrefix, srcString, KeyDelimiter, dstString, KeySuffix)
}

func (nf *NetFlow) Key() NetFlowKey {
	srcString := nf.Src.String(); dstString := nf.Dst.String()
	if srcString < dstString {
		return NetFlowKey(fmt.Sprint(KeyPrefix, srcString, KeyDelimiter, nf.Path.String(), KeyDelimiter, dstString, KeySuffix))
	} else {
		return NetFlowKey(fmt.Sprint(KeyPrefix, dstString, KeyDelimiter, nf.Path.String(), KeyDelimiter, srcString, KeySuffix))
	}
}

func (fl *Flow) Key() FlowKey {
	return FlowKey(fmt.Sprint(KeyPrefix, fl.IPFlow.Key(), KeyDelimiter, fl.NetFlow.Key(), KeyDelimiter, fl.Kind, KeySuffix))
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
	return PacketKey(fmt.Sprint(KeyPrefix, p.Entrypoint.Key(), KeyDelimiter, p.Flow.Key(), KeySuffix))
}