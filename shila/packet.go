package shila

import "fmt"

type Packet struct {
	IP IPPacket
}

type IPPacket struct {
	Raw []byte
}

func (p *Packet) String() string {
	return fmt.Sprint("Size of packet: ", len(p.IP.Raw), " Bytes.")
}
