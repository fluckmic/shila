package layer

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket/layers"
	"shila/core/model"
	"shila/log"
)

const TCPOptionKindMPTCP = 30

type Error string

func (e Error) Error() string {
	return string(e)
}

type MPTCPOption interface{}

type MPTCPOptionBase struct {
	OptionLength  uint8
	OptionSubtype MPTCPOptionSubtype
}

type MPTCPRawOption struct {
	MPTCPOptionBase
	OptionData []byte
}

type MPTCPCapableOptionSender struct {
	MPTCPOptionBase
	Version                uint8
	A, B, C, D, E, F, G, H bool
	SenderKey              uint64
}

type MPTCPCapableOptionSenderReceiver struct {
	MPTCPCapableOptionSender
	ReceiverKey uint64
}

type MPTCPJoinOptionSYN struct {
	MPTCPOptionBase
	B                  bool
	AddressID          uint8
	ReceiverToken      uint32
	SenderRandomNumber uint32
}

type MPTCPJoinOptionSYNACK struct {
	MPTCPOptionBase
	B                  bool
	AddressID          uint8
	SenderTruncHMAC    uint64
	SenderRandomNumber uint32
}

type MPTCPJoinOptionThirdACK struct {
	MPTCPOptionBase
	SenderHMAC []byte
}

type MPTCPOptionSubtype uint8

const (
	MPTCPOptionSubtypeMultipathCapable      = 0 // len = 12 or 20
	MPTCPOptionSubtypeJoinConnection        = 1 // len = 12 (SYN) / 16 (SYN/ACK) / 24 (3rd ACK)
	MPTCPOptionSubtypeDataSequenceSignal    = 2
	MPTCPOptionSubtypeAddAddress            = 3
	MPTCPOptionSubtypeRemoveAddress         = 4
	MPTCPOptionSubtypeChangeSubflowPriority = 5
	MPTCPOptionSubtypeFallback              = 6 // len = 12
	MPTCPOptionSubtypeFastClose             = 7
)

func GetMPTCPReceiverToken(raw []byte) (model.MPTCPEndpointToken, bool, error) {
	// We parse the IPv4 and the TCP layer again. Getting the receiver token is done
	// only once at the setup of a new sub flow. It should be fine to do this twice.
	if _, tcp, err := decodeIPv4andTCPLayer(raw); err != nil {
		// Error in decoding the ipv4/tcp options
		return model.MPTCPEndpointToken(0), false, err
	} else {
		if mptcpOptions, err := decodeMPTCPOptions(tcp); err != nil {
			// MPTCP options does not contain the receiver token
			return model.MPTCPEndpointToken(0), false, nil
		} else {
			for _, mptcpOption := range mptcpOptions {
				if mptcpJoinOptionSYN, ok := mptcpOption.(MPTCPJoinOptionSYN); ok {
					return model.MPTCPEndpointToken(mptcpJoinOptionSYN.ReceiverToken), true, nil
				}
			}
			// Error in decoding the mptcp options
			return model.MPTCPEndpointToken(0), false, err
		}
	}
}

func GetMPTCPSenderKey(raw []byte) (model.MPTCPEndpointKey, bool, error) {
	// We parse the IPv4 and the TCP layer again. Getting the receiver token is done
	// only once at the setup of a new sub flow. It should be fine to do this twice.
	if _, tcp, err := decodeIPv4andTCPLayer(raw); err != nil {
		// Error in decoding the ipv4/tcp options
		return model.MPTCPEndpointKey(0), false, err
	} else {
		if mptcpOptions, err := decodeMPTCPOptions(tcp); err != nil {
			// Error in decoding the mptcp options
			return model.MPTCPEndpointKey(0), false, err
		} else {
			for _, mptcpOption := range mptcpOptions {
				if mptcpCapableOptionSender, ok := mptcpOption.(MPTCPCapableOptionSender); ok {
					return model.MPTCPEndpointKey(mptcpCapableOptionSender.SenderKey), true, nil
				}
			}
			// MPTCP options does not contain the senders key
			return model.MPTCPEndpointKey(0), false, nil
		}
	}
}

func decodeMPTCPOptions(tcp layers.TCP) (options []MPTCPOption, err error) {

	options = []MPTCPOption{}

	// Loop over all options, find the MPTCP options and decode them further.
	for _, option := range tcp.Options {
		if option.OptionType == layers.TCPOptionKind(TCPOptionKindMPTCP) {

			var opt MPTCPOption

			length := option.OptionLength
			subtype := MPTCPOptionSubtype(option.OptionData[0] >> 4)

			optBase := MPTCPOptionBase{length, subtype}
			data := option.OptionData

			switch subtype {

			case MPTCPOptionSubtypeMultipathCapable:

				if length != 12 && length != 20 {
					err = Error(fmt.Sprint("Invalid length ", length, " for MPTCPOptionSubtypeMultipathCapable"))
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

				opt = MPTCPCapableOptionSender{optBase, version,
					A, B, C, D, E, F, G, H, senderKey}

				if length == 20 {
					opt = MPTCPCapableOptionSenderReceiver{opt.(MPTCPCapableOptionSender),
						binary.BigEndian.Uint64(data[10:17])}
				}

			case MPTCPOptionSubtypeJoinConnection:

				switch length {

				case 12:

					B := data[0]&0x1 != 0
					addressID := data[1]

					receiverToken := binary.BigEndian.Uint32(data[2:6])
					senderRandomNumber := binary.BigEndian.Uint32(data[6:10])

					opt = MPTCPJoinOptionSYN{optBase, B, addressID,
						receiverToken, senderRandomNumber}

				case 16:

					B := data[0]&0x1 != 0
					addressID := data[1]

					senderTruncHMAC := binary.BigEndian.Uint64(data[2:10])
					senderRandomNumber := binary.BigEndian.Uint32(data[10:14])

					opt = MPTCPJoinOptionSYNACK{optBase, B, addressID,
						senderTruncHMAC, senderRandomNumber}

				case 24:

					opt = MPTCPJoinOptionThirdACK{optBase, data[2:22]}

				default:
					err = Error(fmt.Sprint("Invalid length ", length, " for  MPTCPOptionSubtypeJoinConnection"))
					return
				}

			case MPTCPOptionSubtypeDataSequenceSignal, MPTCPOptionSubtypeAddAddress, MPTCPOptionSubtypeRemoveAddress,
				MPTCPOptionSubtypeChangeSubflowPriority, MPTCPOptionSubtypeFallback, MPTCPOptionSubtypeFastClose:
				opt = MPTCPRawOption{optBase, data}

			default:
				log.Info.Println("Encountered invalid MPTCPOptionSubtype", subtype, "- Continuing anyway.")
				continue
			}

			options = append(options, opt)
		}
	}

	return
}
