package tcpip

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"shila/shutdown"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

// TODO: ByteOrder!
var hostByteOrder = binary.BigEndian

type IPv4Option interface{}

func DecodeIPv4POptions(ip layers.IPv4) (options []IPv4Option, err error) {
	options = []IPv4Option{}
	return
}

func DecodeSrcAndDstTCPAddr(raw []byte) (net.TCPAddr, net.TCPAddr, error) {
	if ip4v, tcp, err := DecodeIPv4andTCPLayer(raw); err != nil {
		return net.TCPAddr{}, net.TCPAddr{}, err
	} else {
		return net.TCPAddr{IP: ip4v.SrcIP, Port: int(tcp.SrcPort)},
		       net.TCPAddr{IP: ip4v.DstIP, Port: int(tcp.DstPort)},
		       nil
	}
}

// Start slow but correct..
func DecodeIPv4andTCPLayer(raw []byte) (layers.IPv4, layers.TCP, error) {

	ipv4 := layers.IPv4{}
	tcp  := layers.TCP{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ipv4, &tcp)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(raw, &decoded); err != nil {
		if _, ok := err.(*gopacket.UnsupportedLayerType); !ok {
			return ipv4, tcp, nil //TODO!!!
			//return ipv4, tcp, Error(fmt.Sprint("Could not decode IPv4/TCP layer. - ", err.Error()))
		}
	}
	return ipv4, tcp, nil
}

// Returns the next IPv4 frame. Or throws an error if there is an issue.
func PacketizeRawData(ingressRaw chan byte, sizeReadBuffer int) []byte {
	for {
		rawData := make([]byte, 0, sizeReadBuffer)
		b, open := <-ingressRaw
		if b >> 4 == 4 {
			rawData = append(rawData, b)
			// Read 3 more bytes
			for cnt := 0; cnt < 3; cnt++ {
				b, ok := <-ingressRaw; open = open && ok
				rawData = append(rawData, b)
			}
			length := binary.ByteOrder(hostByteOrder).Uint16(rawData[2:4])
			// Load the remaining bytes of the package
			for cnt := 0; cnt < int(length-4); cnt++ {
				b, ok := <-ingressRaw; open = open && ok
				rawData = append(rawData, b)
			}
			return rawData
		} else if b >> 4 == 6 {
			rawData = append(rawData, b)
			// Read 7 more bytes
			for cnt := 0; cnt < 7; cnt++ {
				b, ok := <-ingressRaw; open = open && ok
				rawData = append(rawData, b)
			}
			length := binary.ByteOrder(hostByteOrder).Uint16(rawData[4:6])
			// Discard the remaining 32 byte of the header and the payload
			for cnt := 0; cnt < int(32+length); cnt++ {
				_, ok := <-ingressRaw; open = open && ok
			}
		} else if !open {
			// Ingress was closed in the meantime, return nil.
			return nil
		} else {
			shutdown.Fatal(Error(fmt.Sprint("Error in server traffic packetizer.")))
		}
	}
}