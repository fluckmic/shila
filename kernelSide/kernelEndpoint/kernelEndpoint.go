//
package kernelEndpoint

import (
	"fmt"
	"io"
	"net"
	"shila/config"
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

func (device *Device) Setup() error {

	if device.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}

	// Allocate the vif
	device.vif = vif.New(device.Name, device.Namespace, network.Subnet(device.IP.String()))

	// Setup the vif
	if err := device.vif.Setup(); err != nil {
		_ = device.vif.Teardown()
		return shila.PrependError(err, fmt.Sprint("Unable to start virtual interface."))
	}

	// Turn the vif up
	if err := device.vif.TurnUp(); err != nil {
		_ = device.vif.TurnDown()
		_ = device.vif.Teardown()
		return shila.PrependError(err, fmt.Sprint("Unable to start virtual interface."))
	}

	// Setup the routing
	if err := device.setupRouting(); err != nil {
		_ = device.removeRouting()
		_ = device.vif.TurnDown()
		_ = device.vif.Teardown()
		return shila.PrependError(err, fmt.Sprint("Unable to setup routing."))
	}

	// Allocate the buffers

	device.channels.ingress = make(chan *shila.Packet, config.Config.KernelEndpoint.SizeIngressBuffer)
	device.channels.egress  = make(chan *shila.Packet, config.Config.KernelEndpoint.SizeEgressBuffer)

	device.state.Set(shila.Initialized)
	return nil
}

func (device *Device) Start() error {

	if device.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}

	go device.serveIngress()
	go device.serveEgress()

	device.state.Set(shila.Running)
	return nil
}

func (device *Device) TearDown() error {

	device.state.Set(shila.TornDown)

	// Remove the routing table associated with the kernel endpoint
	err := device.removeRouting()

	// Deallocate the corresponding instance of the interface
	err = device.vif.TurnDown()
	err = device.vif.Teardown()

	// Ingress channel: This kernel endpoint is the only sender, therefore can close it.
	close(device.channels.ingress)

	return err
}

func (device *Device) Role() shila.EndpointRole {
	return device.label
}

func (device *Device) Identifier() string {
	return fmt.Sprint(device.Role(), " (", device.Name,":", device.IP, ")")
}

func (device *Device) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: device.channels.ingress, Egress: device.channels.egress}
}

func (device *Device) setupRouting() error {

	// ip rule add from <dev ip> table <table id>
	args := []string{"rule", "add", "from", device.IP.String(), "table", fmt.Sprint(device.Number)}
	if err := network.Execute(device.Namespace, args...); err != nil {
		return err
	}

	// ip route add table <table id> default dev <dev name> scope link
	args = []string{"route", "add", "table", fmt.Sprint(device.Number), "default", "dev", device.Name, "scope", "link"}
	if err := network.Execute(device.Namespace, args...); err != nil {
		return err
	}

	return nil
}

func (device *Device) removeRouting() error {

	// ip rule del table <table id>
	args := []string{"rule", "del", "table", fmt.Sprint(device.Number)}
	err := network.Execute(device.Namespace, args...)

	// ip route flush table <table id>
	args = []string{"route", "flush", "table", fmt.Sprint(device.Number)}
	err = network.Execute(device.Namespace, args...)

	return err
}

func (device *Device) serveIngress() {

	ingressRaw := make(chan byte, config.Config.KernelEndpoint.SizeRawIngressBuffer)
	go device.packetize(ingressRaw)

	reader := io.Reader(&device.vif)
	storage := make([]byte, config.Config.KernelEndpoint.SizeRawIngressStorage)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, config.Config.KernelEndpoint.ReadSizeRawIngress)
		if err != nil {
			time.Sleep(time.Duration(config.Config.KernelEndpoint.WaitingTimeUntilEscalation) * time.Second)
			if device.state.Not(shila.Running) {
				close(ingressRaw)
				return
			}
			device.endpointIssues <- shila.EndpointIssuePub{
				Issuer: device,
				Error:  ConnectionError("Unable to read data."),
			}
			return
		}
		for _, b := range storage[:nBytesRead] {
			ingressRaw <- b
		}
	}
}

func (device *Device) serveEgress() {
	writer := io.Writer(&device.vif)
	for p := range device.channels.egress {
		_, err := writer.Write(p.Payload)
		if err != nil {
			time.Sleep(time.Duration(config.Config.KernelEndpoint.WaitingTimeUntilEscalation) * time.Second)
			if device.state.Not(shila.Running) {
				return
			}
			device.endpointIssues <- shila.EndpointIssuePub{
				Issuer: device,
				Error:  ConnectionError("Unable to write data."),
			}
			return
		}
	}
}

func (device *Device) packetize(ingressRaw chan byte) {
	for {
		if rawData, err := tcpip.PacketizeRawData(ingressRaw, config.Config.KernelEndpoint.SizeRawIngressStorage); rawData != nil {
			if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
				// We were not able to get the IP flow from the raw data, but there was no issue parsing
				// the raw data. We therefore just drop the packet and hope that the next one is better..
				log.Error.Print(device.Says(fmt.Sprint("Unable to get IP net flow in packetizer. ", err.Error())))
			} else {
				device.channels.ingress <- shila.NewPacket(device, iPHeader, rawData)
			}
		} else {
			if err == nil {
				// All good, ingress raw closed.
				return
			}
			device.endpointIssues <- shila.EndpointIssuePub{
				Issuer: device,
				Error:  shila.PrependError(err, "Error in raw data packetizer."),
			}
			return
		}
	}
}

func (device *Device) Says(str string) string {
	return  fmt.Sprint(device.Identifier(), ": ", str)
}
