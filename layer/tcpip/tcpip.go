//
package tcpip

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"shila/layer"
	"strconv"
)

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

func DecodeIPv4andTCPLayer(raw []byte) (layers.IPv4, layers.TCP, error) {

	ipv4 := layers.IPv4{}
	tcp  := layers.TCP{}

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ipv4, &tcp)
	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(raw, &decoded); err != nil {
		if _, ok := err.(gopacket.UnsupportedLayerType); !ok {
			return ipv4, tcp, err
		}
	}
	return ipv4, tcp, nil
}

// Returns the next IPv4 frame, or an error if unable to parse.
func PacketizeRawData(ingressRaw chan byte, sizeReadBuffer int) ([]byte, error) {
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
			return rawData, nil
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
			return nil, nil
		} else {
			return nil, layer.ParsingError(fmt.Sprint("Unknown IP version ", b >> 4, "."))
		}
	}
}

// <ip>:<port>
func DecodeTCPAddrFromString(addr string) (net.TCPAddr, error) {
	if host, port, err := net.SplitHostPort(addr); err != nil {
		return net.TCPAddr{}, layer.ParsingError(fmt.Sprint("Cannot parse IP ", addr, "."))
	} else {
		IP := net.ParseIP(host)
		Port, err := strconv.Atoi(port)
		if IP == nil {
			return net.TCPAddr{}, layer.ParsingError(fmt.Sprint("Cannot parse IP ", IP, "."))
		} else if err != nil {
			return net.TCPAddr{}, layer.ParsingError(fmt.Sprint("Cannot parse port ", Port, "."))
		} else {
			return net.TCPAddr{IP: IP, Port: Port}, nil
		}
	}
}