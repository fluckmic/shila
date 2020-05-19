package model

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
)

type MPTCPEndpointToken uint32
type MPTCPEndpointKey   uint64

type Mapping struct {
	addressesFromToken 	 map[MPTCPEndpointToken]NetworkConnectionIdentifier
	addressesFromDstIPv4 map[IPAddressPortKey]NetworkConnectionIdentifier
}

func NewMapping() *Mapping {
	return &Mapping{make(map[MPTCPEndpointToken]NetworkConnectionIdentifier),
				   make(map[IPAddressPortKey]NetworkConnectionIdentifier)}
}

func (m Mapping) RetrieveFromMPTCPEndpointToken(token MPTCPEndpointToken) (NetworkConnectionIdentifier, bool) {
	packetHeader, ok := m.addressesFromToken[token]
	return packetHeader, ok
}

func (m Mapping) InsertFromMPTCPEndpointKey(key MPTCPEndpointKey, srcAddr NetworkAddress, dstAddr NetworkAddress, path NetworkPath) error {

	if token, err := mptcpEndpointKeyToToken(key); err != nil {
		return err
	} else {
		if _, ok := m.addressesFromToken[token]; ok {
			return Error(fmt.Sprint("Unable to insert routing entry for key {", key ,"}. - Entry already exists."))
		} else {
			m.addressesFromToken[token] = NetworkConnectionIdentifier{srcAddr, path, dstAddr}
		}
	}
	return nil
}

func (m Mapping) RetrieveFromIPAddressPortKey(key IPAddressPortKey) (NetworkConnectionIdentifier, bool) {
	packetHeader, ok := m.addressesFromDstIPv4[key]
	return packetHeader, ok
}

func (m Mapping) InsertFromIPAddressPortKey(key IPAddressPortKey, srcAddr NetworkAddress, dstAddr NetworkAddress, path NetworkPath) error {

	if _, ok := m.addressesFromDstIPv4[key]; ok {
		return Error(fmt.Sprint("Unable to insert routing entry for destination IPv4 {", key ,"}. - Entry already exists."))
	} else {
		m.addressesFromDstIPv4[key] = NetworkConnectionIdentifier{srcAddr, path, dstAddr}
	}
	return nil
}

func mptcpEndpointKeyToToken(key MPTCPEndpointKey) (MPTCPEndpointToken, error) {

	// The token is used to identify the MPTCP connection and is a cryptographic hash of the receiver's key, as
	// exchanged in the initial MP_CAPABLE handshake (Section 3.1).  In this specification, the tokens presented in
	// this option are generated by the SHA-1 ([4], [15]) algorithm, truncated to the most significant 32 bits.
	// https://tools.ietf.org/html/rfc6824#section-3.1
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, key); err != nil {
		return MPTCPEndpointToken(0), Error(fmt.Sprint("Unable to create token from receiver key {", key ,"}. - ", err.Error()))
	}
	check := sha1.Sum(buf.Bytes())
	return MPTCPEndpointToken(binary.BigEndian.Uint32(check[0:5])), nil
}