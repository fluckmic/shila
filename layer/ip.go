package layer

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"shila/core/model"
	"shila/shutdown"
)

// TODO: ByteOrder!
var hostByteOrder = binary.BigEndian

type IPv4Option interface{}

func DecodeIPv4POptions(ip layers.IPv4) (options []IPv4Option, err error) {
	options = []IPv4Option{}
	return
}

func GetIPHeader(raw []byte) (model.IPHeader, error) {
	if ip4v, tcp, err := decodeIPv4andTCPLayer(raw); err != nil {
		return model.IPHeader{}, err
	} else {
		return model.IPHeader{
			Src: net.TCPAddr{IP: ip4v.SrcIP, Port: int(tcp.SrcPort)},
			Dst: net.TCPAddr{IP: ip4v.DstIP, Port: int(tcp.DstPort)},
		}, nil
	}
}

func GetNetworkHeaderFromIPOptions(raw []byte) (model.NetworkHeader, bool, error) {
	return model.NetworkHeader{}, false, nil
}

// Start slow but correct..
func decodeIPv4andTCPLayer(raw []byte) (layers.IPv4, layers.TCP, error) {

	ipv4 := layers.IPv4{}
	tcp  := layers.TCP{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ipv4, &tcp)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(raw, &decoded); err != nil {
		return ipv4, tcp, Error(fmt.Sprint("Could not decode IPv4/TCP layer. - ", err.Error()))
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
			// Channel was closed in the meantime, return nil.
			return nil
		} else {
			shutdown.Fatal(Error(fmt.Sprint("Error in server traffic packetizer.")))
		}
	}
}