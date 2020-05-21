package shila

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
)

type MPTCPEndpointToken uint32
type MPTCPEndpointKey   uint64

type Mapping struct {
	addressesFromToken 	 map[MPTCPEndpointToken]	NetFlow
	addressesFromDstIPv4 map[IPAddressPortKey]		NetFlow
}

func NewMapping() *Mapping {
	return &Mapping{make(map[MPTCPEndpointToken]NetFlow),
				   make(map[IPAddressPortKey]NetFlow)}
}

func (m Mapping) RetrieveFromMPTCPEndpointToken(token MPTCPEndpointToken) (NetFlow, bool) {
	packetHeader, ok := m.addressesFromToken[token]
	return packetHeader, ok
}

func (m Mapping) InsertFromMPTCPEndpointKey(key MPTCPEndpointKey, flow NetFlow) error {
	if token, err := mptcpEndpointKeyToToken(key); err != nil {
		return Error(fmt.Sprint("Unable to convert from key to token. - ", err.Error()))
	} else {
		if _, ok := m.addressesFromToken[token]; ok {
			return Error(fmt.Sprint("Entry already exists."))
		} else {
			m.addressesFromToken[token] = flow
		}
	}
	return nil
}

func (m Mapping) RetrieveFromIPAddressPortKey(key IPAddressPortKey) (NetFlow, bool) {
	packetHeader, ok := m.addressesFromDstIPv4[key]
	return packetHeader, ok
}

func (m Mapping) InsertFromIPAddressPortKey(key IPAddressPortKey, srcAddr NetworkAddress, dstAddr NetworkAddress, path NetworkPath) error {

	if _, ok := m.addressesFromDstIPv4[key]; ok {
		return Error(fmt.Sprint("Unable to insert routing entry for destination IPv4 {", key ,"}. - Entry already exists."))
	} else {
		m.addressesFromDstIPv4[key] = NetFlow{srcAddr, path, dstAddr}
	}
	return nil
}

func mptcpEndpointKeyToToken(key MPTCPEndpointKey) (MPTCPEndpointToken, error) {
	// The token is used to identify the MPTCP connection and is a cryptographic hash of the receiver's Key, as
	// exchanged in the initial MP_CAPABLE handshake (Section 3.1).  In this specification, the tokens presented in
	// this option are generated by the SHA-1 ([4], [15]) algorithm, truncated to the most significant 32 bits.
	// https://tools.ietf.org/html/rfc6824#section-3.1
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, key); err != nil {
		return MPTCPEndpointToken(0), err
	}
	check := sha1.Sum(buf.Bytes())
	return MPTCPEndpointToken(binary.BigEndian.Uint32(check[0:5])), nil
}