package kernelEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint/vif"
	"shila/kernelSide/network"
	"shila/layer/tcpip"
)

const nameTunDevice = "tun"

type Device struct {
	Number 		uint8
	Name		string
	Namespace	network.Namespace
	IP			net.IP
	channels   	Channels
	vif        	vif.Device
	state   	shila.EntityState
}

type Channels struct {
	ingress    shila.PacketChannel
	egress     shila.PacketChannel
}

func New(number uint8, namespace network.Namespace, ip net.IP) Device {
	return Device{
		Number:		number,
		Name:		fmt.Sprint(nameTunDevice, number),
		Namespace:	namespace,
		IP:			ip,
		state:		shila.NewEntityState(),
	}
}

func (d *Device) Setup() error {

	if d.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}

	// Allocate the vif
	d.vif = vif.New(d.Name, d.Namespace, network.Subnet(d.IP.String()))

	// Setup the vif
	if err := d.vif.Setup(); err != nil {
		_ = d.vif.Teardown()
		return shila.PrependError(err, fmt.Sprint("Unable to start virtual interface."))
	}

	// Turn the vif up
	if err := d.vif.TurnUp(); err != nil {
		_ = d.vif.TurnDown()
		_ = d.vif.Teardown()
		return shila.PrependError(err, fmt.Sprint("Unable to start virtual interface."))
	}

	// Setup the routing
	if err := d.setupRouting(); err != nil {
		_ = d.removeRouting()
		_ = d.vif.TurnDown()
		_ = d.vif.Teardown()
		return shila.PrependError(err, fmt.Sprint("Unable to setup routing."))
	}

	// Allocate the buffers

	d.channels.ingress    = make(chan *shila.Packet, Config.SizeIngressBuff)
	d.channels.egress  	  = make(chan *shila.Packet, Config.SizeEgressBuff)

	d.state.Set(shila.Initialized)
	return nil
}

func (d *Device) Start() error {

	if d.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}

	go d.serveIngress()
	go d.serveEgress()

	d.state.Set(shila.Running)
	return nil
}

func (d *Device) TearDown() error {

	d.state.Set(shila.TornDown)

	// Remove the routing table associated with the kernel endpoint
	err := d.removeRouting()

	// Deallocate the corresponding instance of the interface
	err = d.vif.TurnDown()
	err = d.vif.Teardown()

	close(d.channels.ingress)
	close(d.channels.egress)

	return err
}

func (d *Device) Label() shila.EndpointLabel {
	return shila.KernelEndpoint
}

func (d *Device) Key() shila.EndpointKey {
	return shila.EndpointKey(shila.GetIPAddressKey(d.IP))
}

func (d *Device) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: d.channels.ingress, Egress: d.channels.egress}
}

func (d *Device) setupRouting() error {

	// ip rule add from <dev ip> table <table id>
	args := []string{"rule", "add", "from", d.IP.String(), "table", fmt.Sprint(d.Number)}
	if err := network.Execute(d.Namespace, args...); err != nil {
		return err
	}

	// ip route add table <table id> default dev <dev name> scope link
	args = []string{"route", "add", "table", fmt.Sprint(d.Number), "default", "dev", d.Name, "scope", "link"}
	if err := network.Execute(d.Namespace, args...); err != nil {
		return err
	}

	return nil
}

func (d *Device) removeRouting() error {

	// ip rule del table <table id>
	args := []string{"rule", "del", "table", fmt.Sprint(d.Number)}
	err := network.Execute(d.Namespace, args...)

	// ip route flush table <table id>
	args = []string{"route", "flush", "table", fmt.Sprint(d.Number)}
	err = network.Execute(d.Namespace, args...)

	return err
}

func (d *Device) serveIngress() {

	ingressRaw := make(chan byte, Config.SizeReadBuffer)
	go d.packetize(ingressRaw)

	reader := io.Reader(&d.vif)
	storage := make([]byte, Config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.BatchSizeRead)
		if err != nil && d.state.Not(shila.Running) {
			// Error doesn't matter, kernel endpoint
			// is no longer valid anyway.
			close(ingressRaw)
			return
		} else if err != nil {
			panic("Handle error in go routine..") //TODO!
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (d *Device) serveEgress() {
	writer := io.Writer(&d.vif)
	for p := range d.channels.egress {
		_, err := writer.Write(p.Payload)
		if err != nil && !d.state.Not(shila.Running) {
			// Error doesn't matter, kernel endpoint
			// is no longer valid anyway.
			return
		} else if err != nil {
			panic("Handle error in go routine..") //TODO!
		}
	}
}

func (d *Device) packetize(ingressRaw chan byte) {
	for {
		if rawData, _ := tcpip.PacketizeRawData(ingressRaw, Config.SizeReadBuffer); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				panic("Handle error in go routine..") //TODO!
				/* panic(fmt.Sprint("Unable to get IP header in packetizer of kernel endpoint {",
				   d.Key(), "}. - ", err.Error())) */
			} else {
				d.channels.ingress <- shila.NewPacket(d, iPHeader, rawData)
			}
		} else {
			// ingress raw closed
			return
		}
	}
}
