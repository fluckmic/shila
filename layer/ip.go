package layer

import "github.com/google/gopacket/layers"

type IPv4Option interface{}

func DecodeIPv4POptions(ip layers.IPv4) (options []IPv4Option, err error) {
	options = []IPv4Option{}
	return
}
