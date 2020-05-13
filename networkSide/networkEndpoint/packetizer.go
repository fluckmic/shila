package networkEndpoint

import (
	"encoding/binary"
	"fmt"
	"shila/core/model"
	"shila/shutdown"
)

// TODO: ByteOrder!
var hostByteOrder = binary.BigEndian

func (c *Client) packetize() {

	// Fatal error could occur.. :o
	shutdown.Check()

	for {
		rawData := make([]byte, 0, c.config.SizeReadBuffer)

		b := <-c.ingressRaw
		switch b >> 4 {
		case 4:
			rawData = append(rawData, b)
			c.ip4(rawData)
		default:
			shutdown.Fatal(Error(fmt.Sprint("Unknown IP version in client packetizer.")))
		}
	}
}

func (c *Client) ip4(storage []byte) {

	// Read 3 more bytes
	for cnt := 0; cnt < 3; cnt++ {
		storage = append(storage, <-c.ingressRaw)
	}

	length := binary.ByteOrder(hostByteOrder).Uint16(storage[2:4])

	// Load the remaining bytes of the package
	for cnt := 0; cnt < int(length-4); cnt++ {
		storage = append(storage, <-c.ingressRaw)
	}

	c.trafficChannels.Ingress <- model.NewPacketFromRawIP(c, storage)
}

