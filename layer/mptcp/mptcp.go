package mptcp

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket/layers"
	"shila/layer/tcpip"
)

const TCPOptionKindMPTCP = 30

type Error string

func (e Error) Error() string {
	return string(e)
}

type Option interface{}

type OptionBase struct {
	OptionLength  uint8
	OptionSubtype OptionSubtype
}

type RawOption struct {
	OptionBase
	OptionData []byte
}

type CapableOptionSender struct {
	OptionBase
	Version                uint8
	A, B, C, D, E, F, G, H bool
	SenderKey              uint64
}

type CapableOptionSenderReceiver struct {
	CapableOptionSender
	ReceiverKey uint64
}

type JoinOptionSYN struct {
	OptionBase
	B                  bool
	AddressID          uint8
	ReceiverToken      uint32
	SenderRandomNumber uint32
}

type JoinOptionSYNACK struct {
	OptionBase
	B                  bool
	AddressID          uint8
	SenderTruncHMAC    uint64
	SenderRandomNumber uint32
}

type JoinOptionThirdACK struct {
	OptionBase
	SenderHMAC []byte
}

type OptionSubtype uint8

type EndpointToken uint32
type EndpointKey uint64

const (
	MultipathCapable      OptionSubtype = 0 // len = 12 or 20
	JoinConnection        OptionSubtype = 1 // len = 12 (SYN) / 16 (SYN/ACK) / 24 (3rd ACK)
	DataSequenceSignal    OptionSubtype = 2
	AddAddress            OptionSubtype = 3
	RemoveAddress         OptionSubtype = 4
	ChangeSubflowPriority OptionSubtype = 5
	Fallback              OptionSubtype = 6 // len = 12
	FastClose             OptionSubtype = 7
)

func (os OptionSubtype) String() string {
	switch os {
	case MultipathCapable    	: return "MultipathCapable"
	case JoinConnection      	: return "JoinConnection"
	case DataSequenceSignal  	: return "DataSequenceSignal"
	case AddAddress          	: return "AddAddress"
	case RemoveAddress       	: return "RemoveAddress"
	case ChangeSubflowPriority 	: return "ChangeSubflowPriority"
	case Fallback               : return "Fallback"
	case FastClose              : return "FastClose"
	}
	return "Unknown"
}

func GetReceiverToken(raw []byte) (EndpointToken, bool, error) {
	if _, tcp, err := tcpip.DecodeIPv4andTCPLayer(raw); err != nil {
		// Error in decoding the ipv4/tcp options
		return EndpointToken(0), false, err
	} else {
		if mptcpOptions, err := decodeMPTCPOptions(tcp); err != nil {
			// MPTCP options does not contain the receiver token
			return EndpointToken(0), false, nil
		} else {
			for _, mptcpOption := range mptcpOptions {
				if mptcpJoinOptionSYN, ok := mptcpOption.(JoinOptionSYN); ok {
					return EndpointToken(mptcpJoinOptionSYN.ReceiverToken), true, nil
				}
			}
			// Error in decoding the mptcp options
			return EndpointToken(0), false, err
		}
	}
}

func GetSenderKey(raw []byte) (EndpointKey, bool, error) {
	if _, tcp, err := tcpip.DecodeIPv4andTCPLayer(raw); err != nil {
		// Error in decoding the ipv4/tcp options
		return EndpointKey(0), false, err
	} else {
		if mptcpOptions, err := decodeMPTCPOptions(tcp); err != nil {
			// Error in decoding the mptcp options
			return EndpointKey(0), false, err
		} else {
			for _, mptcpOption := range mptcpOptions {
				if mptcpCapableOptionSender, ok := mptcpOption.(CapableOptionSender); ok {
					return EndpointKey(mptcpCapableOptionSender.SenderKey), true, nil
				}
			}
			// MPTCP options does not contain the senders key
			return EndpointKey(0), false, nil
		}
	}
}

func EndpointKeyToToken(key EndpointKey) (EndpointToken, error) {
	// The token is used to identify the MPTCP connection and is a cryptographic hash of the receiver's Key, as
	// exchanged in the initial MP_CAPABLE handshake (Section 3.1).  In this specification, the tokens presented in
	// this option are generated by the SHA-1 algorithm, truncated to the most significant 32 bits.
	// https://tools.ietf.org/html/rfc6824#section-3.1
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, key); err != nil {
		return EndpointToken(0), err
	}
	check := sha1.Sum(buf.Bytes())
	return EndpointToken(binary.BigEndian.Uint32(check[0:5])), nil
}

func decodeMPTCPOptions(tcp layers.TCP) (options []Option, err error) {

	options = []Option{}

	// Loop over all options, find the MPTCP options and decode them further.
	for _, option := range tcp.Options {
		if option.OptionType == layers.TCPOptionKind(TCPOptionKindMPTCP) {

			var opt Option

			length := option.OptionLength
			subtype := OptionSubtype(option.OptionData[0] >> 4)

			optBase := OptionBase{OptionLength: length, OptionSubtype: subtype}
			data := option.OptionData

			switch subtype {

			case MultipathCapable:

				if length != 12 && length != 20 {
					err = Error(fmt.Sprint("Invalid length {", length, "} for {", MultipathCapable, "}."))
					return
				}

				version := data[0] & 0xf

				A := data[1]&0x80 != 0
				B := data[1]&0x40 != 0
				C := data[1]&0x20 != 0
				D := data[1]&0x10 != 0
				E := data[1]&0x8 != 0
				F := data[1]&0x4 != 0
				G := data[1]&0x2 != 0
				H := data[1]&0x1 != 0

				senderKey := binary.BigEndian.Uint64(data[2:10])

				opt = CapableOptionSender{
					OptionBase: optBase,
					Version: 	version,
					A: A, B: B, C: C, D: D,
					E: E, F: F, G: G, H: H,
					SenderKey: 	senderKey,
				}

				if length == 20 {
					opt = CapableOptionSenderReceiver{
						CapableOptionSender: opt.(CapableOptionSender),
						ReceiverKey:		 binary.BigEndian.Uint64(data[10:17]),
					}
				}

			case JoinConnection:

				switch length {

				case 12:

					B := data[0]&0x1 != 0
					addressID := data[1]

					receiverToken := binary.BigEndian.Uint32(data[2:6])
					senderRandomNumber := binary.BigEndian.Uint32(data[6:10])

					opt = JoinOptionSYN{
						OptionBase: 		optBase,
						B: 					B,
						AddressID: 			addressID,
						ReceiverToken: 		receiverToken,
						SenderRandomNumber: senderRandomNumber,
					}

				case 16:

					B := data[0]&0x1 != 0
					addressID := data[1]

					senderTruncHMAC := binary.BigEndian.Uint64(data[2:10])
					senderRandomNumber := binary.BigEndian.Uint32(data[10:14])

					opt = JoinOptionSYNACK{
						OptionBase: 		optBase,
						B: 					B,
						AddressID: 			addressID,
						SenderTruncHMAC: 	senderTruncHMAC,
						SenderRandomNumber: senderRandomNumber,
					}

				case 24:

					opt = JoinOptionThirdACK{OptionBase: optBase, SenderHMAC: data[2:22]}

				default:
					err = Error(fmt.Sprint("Invalid length {", length, "} for {", JoinConnection, "}."))
					return
				}

			case DataSequenceSignal, AddAddress, RemoveAddress,
				ChangeSubflowPriority, Fallback, FastClose:
				opt = RawOption{OptionBase: optBase, OptionData: data}

			default:
				continue
			}

			options = append(options, opt)
		}
	}

	return
}