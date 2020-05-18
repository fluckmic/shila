package model

import (
	"fmt"
	"net"
)

// Keys

// (ipv4)
type IPAddressKey 			  string
// (ipv4:port)
type IPAddressPortKey		  string
// (src-ipv4-address:port<>dst-ipv4-address:port)
type IPHeaderKey   			  string
// (network-address)
type NetworkAddressKey		  string
// (network-address<>path)
type NetworkAddressAndPathKey string
// (src-network-address<>path<>dst-network-address)
type NetworkHeaderKey 		  string

type KeyGenerator struct {}

func (g KeyGenerator) IPAddressKey(ip net.IP) IPAddressKey {
	return IPAddressKey(fmt.Sprint("(",ip.String(),")"))
}

func (g KeyGenerator) IPAddressPortKey(addr net.TCPAddr) IPAddressPortKey {
	return IPAddressPortKey(fmt.Sprint("(", addr.String(), ")"))
}

func (g KeyGenerator) IPHeaderKey(iph IPHeader) IPHeaderKey {
	srcString := iph.Src.String(); dstString := iph.Dst.String()
	if srcString < dstString {
		return IPHeaderKey(fmt.Sprint("(", srcString, "<>", dstString, ")"))
	} else {
		return IPHeaderKey(fmt.Sprint("(", dstString, "<>", srcString, ")"))
	}
}

func (g KeyGenerator) NetworkHeaderKey(nh NetworkHeader) NetworkHeaderKey {
	return NetworkHeaderKey(fmt.Sprint("(", nh.Src.String(), "<>", nh.Path.String(), "<>", nh.Dst.String(), ")"))
}

func (g KeyGenerator) NetworkAddressAndPathKey(address NetworkAddress, path NetworkPath) NetworkAddressAndPathKey {
	return NetworkAddressAndPathKey(fmt.Sprint("(",address.String(),"<>",path.String(),")"))
}
