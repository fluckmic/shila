package layer

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket/layers"
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

func DecodeMPTCPOptions(tcp layers.TCP) (options []MPTCPOption, err error) {

	options = []MPTCPOption{}

	// Loop over all options, find the MPTCP options and decode them further.
	for _, option := range tcp.Options {
		if option.OptionType == layers.TCPOptionKind(TCPOptionKindMPTCP) {

			optLength := option.OptionLength
			optSubtype := MPTCPOptionSubtype(option.OptionData[0] >> 4)

			optBase := MPTCPOptionBase{optLength, optSubtype}
			optData := option.OptionData

			switch optSubtype {

			case MPTCPOptionSubtypeMultipathCapable:

				switch optLength {
				case 12:
					continue
				case 20:
					continue
				default:
					continue
				}

			case MPTCPOptionSubtypeJoinConnection:

				switch optLength {
				case 12:
					continue
				case 20:
					continue
				default:
					continue
				}

				continue

			case MPTCPOptionSubtypeDataSequenceSignal, MPTCPOptionSubtypeAddAddress, MPTCPOptionSubtypeRemoveAddress,
				MPTCPOptionSubtypeChangeSubflowPriority, MPTCPOptionSubtypeFallback, MPTCPOptionSubtypeFastClose:
				options = append(options, MPTCPRawOption{optBase, optData})
				continue

			default:
				log.Info.Println("Encountered invalid MPTCPOptionSubtype", optSubtype, "- Continuing anyway.")
				continue
			}
		}
	}

	return
}

/*
func (mptcp *MPTCPCapableOption) DecodeFromMPTCPOption(opt layers.TCPOption) error {

	data := opt.OptionData

	mptcp.OptionLength  = opt.OptionLength

	mptcp.OptionSubtype = data[0] >> 4
	if mptcp.OptionSubtype != MPTCPOptionSubtypeMultipathCapable {
		return Error(fmt.Sprint("Invalid MPTCP option",
			" - ", "Unknown MPTCP subtype ", mptcp.OptionSubtype, "." ))
	}

	mptcp.OptionSubtype = opt.OptionSubtype

	mptcp.Version 		= data[0] & 0xf

	mptcp.A				= data[1] & 0x80 != 0
	mptcp.B 			= data[1] & 0x40 != 0
	mptcp.C 			= data[1] & 0x20 != 0
	mptcp.D 			= data[1] & 0x10 != 0
	mptcp.E 			= data[1] & 0x8  != 0
	mptcp.F 			= data[1] & 0x4  != 0
	mptcp.G 			= data[1] & 0x2  != 0
	mptcp.H 			= data[1] & 0x1  != 0

	mptcp.SenderKey 	= binary.BigEndian.Uint64(data[2:10])

	if mptcp.OptionLength == 20 {
		mptcp.ReceiverKey = binary.BigEndian.Uint64(data[10:17])
	}

	return nil
}

func (mptcp *MPTCPJoinOptionSYN) DecodeFromMPTCPOption(opt MPTCPOption) error {

	data := opt.OptionData

	mptcp.OptionLength  		= opt.OptionLength
	mptcp.OptionSubtype 		= opt.OptionSubtype

	mptcp.B						= data[0] & 0x1  != 0
	mptcp.AddressID 			= data[1]

	mptcp.ReceiverToken 		= binary.BigEndian.Uint32(data[2:6])
	mptcp.SenderRandomNumber	= binary.BigEndian.Uint32(data[6:10])

	return nil
}

func (mptcp *MPTCPJoinOptionSYNACK) DecodeFromMPTCPOption(opt MPTCPOption) error {

	data := opt.OptionData

	mptcp.OptionLength  		= opt.OptionLength
	mptcp.OptionSubtype 		= opt.OptionSubtype

	mptcp.B						= data[0] & 0x1  != 0
	mptcp.AddressID 			= data[1]

	mptcp.SenderTruncHMAC 		= binary.BigEndian.Uint64(data[2:10])
	mptcp.SenderRandomNumber	= binary.BigEndian.Uint32(data[10:14])

	return nil
}

func (mptcp *MPTCPJoinOptionThirdACK) DecodeFromMPTCPOption(opt MPTCPOption) error {

	mptcp.OptionLength  = opt.OptionLength
	mptcp.OptionSubtype = opt.OptionSubtype

	mptcp.SenderHMAC 			=  opt.OptionData[2:22]

	return nil
}

func ()

*/
