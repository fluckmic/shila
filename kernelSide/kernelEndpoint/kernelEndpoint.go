// TODO: Add a more detailed description.
// application endpoint
package kernelEndpoint

import (
	"fmt"
	"io"
	"shila/config"
	"shila/core/shila"
	"shila/helper"
	"shila/kernelSide/kernelEndpoint/vif"
	"shila/layer/tcpip"
	"shila/log"
)

const doVerbosePackageLogging = false

func verbolog(loggingMessage string) {
	if doVerbosePackageLogging {
		log.Verbose.Print(loggingMessage)
	}
}

type Device struct {
	Id         Identifier
	channels   Channels
	config     config.KernelEndpoint
	packetizer *Device
	vif        *vif.Device
	isValid    bool
	isSetup    bool // TODO: merge to "state" object
	isRunning  bool
}

type Channels struct {
	ingressRaw chan byte
	ingress    shila.PacketChannel
	egress     shila.PacketChannel
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func New(id Identifier, config config.KernelEndpoint) *Device {
	return &Device{
		Id: 		id,
		config: 	config,
		isValid: 	true,
	}
}

func (d *Device) Setup() error {

	if !d.IsValid() {
		return Error(fmt.Sprint("Unable to setup kernel endpoint ",
			d.Id.Name(), " - ", "Device no longer valid."))
	}

	if d.IsRunning() {
		return Error(fmt.Sprint("Unable to setup kernel endpoint ",
			d.Id.Name(), " - ", "Device already running."))
	}

	if d.IsSetup() {
		return nil
	}

	// Allocate the vif
	d.vif = vif.New(d.Id.Name(), d.Id.namespace, d.Id.IP())

	// Setup the vif
	if err := d.vif.Setup(); err != nil {
		_ = d.vif.Teardown()
		d.vif = nil
		return err
	}

	// Turn the vif up
	if err := d.vif.TurnUp(); err != nil {
		_ = d.vif.TurnDown()
		_ = d.vif.Teardown()
		d.vif = nil
		return err
	}

	// Setup the routing
	if err := d.setupRouting(); err != nil {
		_ = d.removeRouting()
		_ = d.vif.TurnDown()
		_ = d.vif.Teardown()
		d.vif = nil
		return err
	}

	// Allocate the buffers
	d.channels.ingressRaw = make(chan byte, d.config.SizeReadBuffer)
	d.channels.ingress    = make(chan *shila.Packet, d.config.SizeIngressBuff)
	d.channels.egress  	  = make(chan *shila.Packet, d.config.SizeEgressBuff)

	d.isSetup = true
	return nil
}

func (d *Device) TearDown() error {

	d.isValid = false
	d.isRunning = false
	d.isSetup = false

	// Remove the routing table associated with the kernel endpoint
	err := d.removeRouting()

	// Deallocate the corresponding instance of the interface
	err = d.vif.TurnDown()
	err = d.vif.Teardown()

	d.vif = nil
	d.channels.ingressRaw 	= nil
	d.channels.ingress 		= nil
	d.channels.egress 		= nil

	return err
}

func (d *Device) Start() error {

	if !d.IsValid() {
		return Error(fmt.Sprint("Cannot start kernel endpoint ",
			d.Id.Name(), " - ", "Device no longer valid."))
	}

	if !d.IsSetup() {
		return Error(fmt.Sprint("Cannot start kernel endpoint ",
			d.Id.Name(), " - ", "Device not yet setup."))
	}

	if d.IsRunning() {
		return Error(fmt.Sprint("Cannot start kernel endpoint ",
			d.Id.Name(), " - ", "Device already running."))

	}

	go d.packetize()
	go d.serveIngress()
	go d.serveEgress()

	d.isRunning = true

	return nil
}

func (d *Device) IsValid() bool {
	return d.isValid
}

func (d *Device) IsSetup() bool {
	return d.isSetup
}

func (d *Device) IsRunning() bool {
	return d.isRunning
}

func (d *Device) setupRouting() error {

	// ip rule add from <dev ip> table <table id>
	args := []string{"rule", "add", "from", d.Id.IP(), "table", fmt.Sprint(d.Id.Number())}
	if err := helper.ExecuteIpCommand(d.Id.namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", err.Error()))
	}

	// ip route add table <table id> default dev <dev name> scope link
	args = []string{"route", "add", "table", fmt.Sprint(d.Id.Number()), "default", "dev", d.Id.Name(), "scope", "link"}
	if err := helper.ExecuteIpCommand(d.Id.namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", err.Error()))
	}

	return nil
}

func (d *Device) removeRouting() error {

	// ip rule del table <table id>
	args := []string{"rule", "del", "table", fmt.Sprint(d.Id.number)}
	err := helper.ExecuteIpCommand(d.Id.namespace, args...)

	// ip route flush table <table id>
	args = []string{"route", "flush", "table", fmt.Sprint(d.Id.number)}
	err = helper.ExecuteIpCommand(d.Id.namespace, args...)

	return err
}

func (d *Device) serveIngress() {
	reader := io.Reader(d.vif)
	storage := make([]byte, d.config.SizeReadBuffer)
	for {
		nBytesRead, err := io.ReadAtLeast(reader, storage, d.config.BatchSizeRead)
		verbolog(fmt.Sprint("Kernel endpoint {", d.Key() ,"} has read {", nBytesRead, "} " +
			"bytes which are now send to the ingress raw."))
		if err != nil && !d.IsValid() {
			// Error doesn't matter, kernel endpoint is no longer valid anyway.
			return
		} else if err != nil {
			panic("implement me") //TODO!
		}
		for _, b := range storage[:nBytesRead] {
			d.channels.ingressRaw <- b
		}
	}
}

func (d *Device) serveEgress() {
	writer := io.Writer(d.vif)
	for p := range d.channels.egress {
		_, err := writer.Write(p.Payload)
		if err != nil && !d.IsValid() {
			// Error doesn't matter, kernel endpoint is no longer valid anyway.
			return
		} else if err != nil {
			panic("implement me") //TODO!
		}
	}
}

func (d *Device) packetize() {
	for {
		rawData, _ := tcpip.PacketizeRawData(d.channels.ingressRaw, d.config.SizeReadBuffer)		// TODO: Handle error
		if iPHeader, err := shila.GetIPFlow(rawData); err != nil {
			panic(fmt.Sprint("Unable to get IP header in packetizer of kernel endpoint {", d.Key(),
				"}. - ", err.Error())) // TODO: Handle panic!
		} else {
			verbolog(fmt.Sprint("Kernel endpoint {", d.Id.Key() ,"} enqueued new ingress packet."))
			d.channels.ingress <- shila.NewPacket(d, iPHeader, rawData)
		}
	}
}

func (d *Device) Label() shila.EndpointLabel {
	return shila.KernelEndpoint
}

// TODO:
func (d *Device) Key() shila.EndpointKey {
	return shila.EndpointKey(d.Id.Key())
}

func (d *Device) TrafficChannels() shila.PacketChannels {
	return shila.PacketChannels{Ingress: d.channels.ingress, Egress: d.channels.egress}
}