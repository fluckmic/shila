package kerep

import (
	"encoding/binary"
	"fmt"
	"shila/shila"
	"shila/shutdown"
)

// TODO: ByteOrder!
var hostByteOrder = binary.BigEndian

func (d *Device) packetize() {

	// Fatal error could occur.. :o
	shutdown.Check()

	for {
		rawData := make([]byte, 0, d.config.SizeReadBuffer)

		b := <-d.Channels.ingressRaw
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
		storage = append(storage, <-d.Channels.ingressRaw)
	}

	length := binary.ByteOrder(hostByteOrder).Uint16(storage[2:4])

	// Load the remaining bytes of the package
	for cnt := 0; cnt < int(length-4); cnt++ {
		storage = append(storage, <-d.Channels.ingressRaw)
	}

	d.Channels.Ingress <- shila.NewPacketFromRawIP(d, storage)
}

func (d *Device) ip6(storage []byte) {

	// Read 7 more bytes
	for cnt := 0; cnt < 7; cnt++ {
		storage = append(storage, <-d.Channels.ingressRaw)
	}

	length := binary.ByteOrder(hostByteOrder).Uint16(storage[4:6])

	// Discard the remaining 32 byte of the header and the payload
	for cnt := 0; cnt < int(32+length); cnt++ {
		<-d.Channels.ingressRaw
	}
}
