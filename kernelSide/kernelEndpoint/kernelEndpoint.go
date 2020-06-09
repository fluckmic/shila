//
package kernelEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint/vif"
	"shila/kernelSide/network"
	"shila/layer/tcpip"
	"shila/log"
	"time"
)

const nameTunDevice = "tun"

type Device struct {
	Number 			uint8
	Name			string
	Namespace		network.Namespace
	IP				net.IP
	label 			shila.EndpointRole
	endpointIssues 	shila.EndpointIssuePubChannel
	channels   		Channels
	vif        		vif.Device
	state   		shila.EntityState
}

type Channels struct {
	ingress    shila.PacketChannel
	egress     shila.PacketChannel
}

func New(number uint8, namespace network.Namespace, ip net.IP, label shila.EndpointRole, endpointIssues shila.EndpointIssuePubChannel) Device {
	return Device{
		Number:			number,
		Name:			fmt.Sprint(nameTunDevice, number),
		Namespace:		namespace,
		IP:				ip,
		label:			label,
		endpointIssues: endpointIssues,
		state:			shila.NewEntityState(),
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

	d.channels.ingress    = make(chan *shila.Packet, Config.SizeIngressBuffer)
	d.channels.egress  	  = make(chan *shila.Packet, Config.SizeEgressBuffer)

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

	// Ingress channel: This kernel endpoint is the only sender, therefore can close it.
	close(d.channels.ingress)

	return err
}

func (d *Device) Role() shila.EndpointRole {
	return d.label
}

func (d *Device) Identifier() shila.EndpointIdentifier {
	return shila.EndpointIdentifier(shila.GetIPAddressKey(d.IP))
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

	ingressRaw := make(chan byte, Config.SizeRawIngressBuffer)
	go d.packetize(ingressRaw)

	reader := io.Reader(&d.vif)
	storage := make([]byte, Config.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, Config.ReadSizeRawIngress)
		if err != nil {
			time.Sleep(Config.WaitingTimeUntilEscalation)
			if d.state.Not(shila.Running) {
				close(ingressRaw)
				return
			}
			d.endpointIssues <- shila.EndpointIssuePub{
				Issuer: d,
				Error:  shila.ThirdPartyError("Unable to read data."),
			}
			return
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
		if err != nil {
			time.Sleep(Config.WaitingTimeUntilEscalation)
			if d.state.Not(shila.Running) {
				return
			}
			d.endpointIssues <- shila.EndpointIssuePub{
				Issuer: d,
				Error:  shila.ThirdPartyError("Unable to write data."),
			}
			return
		}
	}
}

func (d *Device) packetize(ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, Config.SizeRawIngressStorage); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP flow from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print("Unable to get IP net flow in packetizer of kernel endpoint {", d.Identifier(), "}. - ", err.Error())
			} else {
				d.channels.ingress <- shila.NewPacket(d, iPHeader, rawData)
			}
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			d.endpointIssues <- shila.EndpointIssuePub{
				Issuer: d,
				Error:  shila.PrependError(err, "Error in raw data packetizer."),
			}
			return
		}
	}
}

func (d *Device) Says(string) string {
	panic("implement me.")
}
