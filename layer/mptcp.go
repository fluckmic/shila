package layer

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket/layers"
)

const TCPOptionKindMPTCP = 30

type Error string

func (e Error) Error() string {
	return string(e)
}

type MPTCPOption struct {
	OptionLength  			uint8
	OptionSubtype 			MPTCPOptionSubtype
	OptionData    			[]byte
}

type MPTCPCapableOption struct {
	OptionLength  			uint8
	OptionSubtype 			MPTCPOptionSubtype
	Version		  			uint8
	A, B, C, D, E, F, G, H	bool
	SenderKey				uint64
	ReceiverKey				uint64
}

type MPTCPJoinOptionSYN struct {
	OptionLength  			uint8
	OptionSubtype 			MPTCPOptionSubtype
	B 						bool
	AddressID				uint8
	ReceiverToken			uint32
	SenderRandomNumber		uint32
}

type MPTCPJoinOptionSYNACK struct {
	OptionLength  			uint8
	OptionSubtype 			MPTCPOptionSubtype
	B 						bool
	AddressID				uint8
	SenderTruncHMAC			uint64
	SenderRandomNumber		uint32
}

type MPTCPJoinOptionThirdACK struct {
	OptionLength  			uint8
	OptionSubtype 			MPTCPOptionSubtype
	B 						bool
	SenderHMAC				[20]byte
}

// MPTCPOptionSubtype represents the MPTCP option subtype code.
type MPTCPOptionSubtype uint8

const (
	MPTCPOptionSubtypeMultipathCapable             	= 0  // len = 12 or 20
	MPTCPOptionSubtypeJoinConnection               	= 1  // len = 12 (SYN) / 16 (SYN/ACK) / 24 (3rd ACK)
	MPTCPOptionSubtypeDataSequenceSignal           	= 2
	MPTCPOptionSubtypeAddAddress                   	= 3
	MPTCPOptionSubtypeRemoveAddress                	= 4
	MPTCPOptionSubtypeChangeSubflowPriority		   	= 5
	MPTCPOptionSubtypeFallback					   	= 6  // len = 12
	MPTCPOptionSubtypeFastClose						= 7
	MPTCPOptionInvalid								= 8
)

func (mptcp *MPTCPOption) DecodeFromTCPOption(opt layers.TCPOption) error {

	mptcp.OptionLength = opt.OptionLength

	mptcp.OptionSubtype = MPTCPOptionSubtype(opt.OptionData[0] >> 4)
	if mptcp.OptionSubtype >= MPTCPOptionInvalid {
		return Error(fmt.Sprint("Invalid MPTCP option",
			" - ", "Unknown MPTCP subtype ", mptcp.OptionSubtype, "." ))
	}

	mptcp.OptionData   = opt.OptionData
	return nil
}

func (mptcp *MPTCPCapableOption) DecodeFromMPTCPOption(opt MPTCPOption) error {

	data := opt.OptionData

	mptcp.OptionLength  = opt.OptionLength
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

	// TODO: Include a ReceiverKeyValid bool?
	if mptcp.OptionLength == 20 {
		mptcp.ReceiverKey = binary.BigEndian.Uint64(data[10:17])
	}

	return nil
}

func (mptcp *MPTCPJoinOptionSYN) DecodeFromMPTCPOption(opt MPTCPOption) error {
	return nil
}

func (mptcp *MPTCPJoinOptionSYNACK) DecodeFromMPTCPOption(opt MPTCPOption) error {
	return nil
}

func (mptcp *MPTCPJoinOptionThirdACK) DecodeFromMPTCPOption(opt MPTCPOption) error {
	return nil
}