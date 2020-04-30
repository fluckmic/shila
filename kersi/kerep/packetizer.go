package packetizer

import (
	"encoding/binary"
	"fmt"
	"shila/shila"
	"shila/shutdown"
)

// TODO: ByteOrder!
var hostByteOrder = binary.BigEndian

type Device struct {
	//kerep 	   *kerep.Device
	input      chan byte
	output     chan *shila.Packet
	bufferSize int
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func New(in chan byte, out chan *shila.Packet, bufferSize int) *Device {
	return &Device{in, out, bufferSize}
}

func (d *Device) Run() {

	// Fatal error could occur.. :o
	shutdown.Check()

	for {
		rawData := make([]byte, 0, d.bufferSize)

		b := <-d.input
		switch b >> 4 {
		case 4:
			rawData = append(rawData, b)
			d.ip4(rawData)
		case 6:
			rawData = append(rawData, b)
			d.ip6(rawData)
		default:
			shutdown.Fatal(Error(fmt.Sprint("Unknown IPv in packetizer.")))
		}
	}
}

func (d *Device) ip4(storage []byte) {

	// Read 3 more bytes
	for cnt := 0; cnt < 3; cnt++ {
		storage = append(storage, <-d.input)
	}

	length := binary.ByteOrder(hostByteOrder).Uint16(storage[2:4])

	// Load the remaining bytes of the package
	for cnt := 0; cnt < int(length-4); cnt++ {
		storage = append(storage, <-d.input)
	}

	//d.output <- shila.NewPacketFromRawIP(d.kerep, storage)
}

func (d *Device) ip6(storage []byte) {

	// Read 7 more bytes
	for cnt := 0; cnt < 7; cnt++ {
		storage = append(storage, <-d.input)
	}

	length := binary.ByteOrder(hostByteOrder).Uint16(storage[4:6])

	// Discard the remaining 32 byte of the header and the payload
	for cnt := 0; cnt < int(32+length); cnt++ {
		<-d.input
	}
}
