// TODO: Add a more detailed description.
// application endpoint
package kerep

import (
	"fmt"
	"shila/config"
	"shila/helper"
	"shila/kersi/kerep/vif"
	"shila/shila"
)

type Device struct {
	Id        Identifier
	Buffers   *Buffer
	config    config.KernelEndpoint
	vif       *vif.Device
	isRunning bool
}

type Buffer struct {
	Ingress chan *shila.Packet
	Egress  chan *shila.Packet
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func New(id Identifier, config config.KernelEndpoint) *Device {
	var buf = Buffer{nil, nil}
	return &Device{id, &buf, config, nil, false}
}

func (d *Device) Setup() error {

	if d.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel endpoint ",
			d.Id.Name(), " - ", "Device already setup."))
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
	d.Buffers.Ingress = make(chan *shila.Packet, d.config.SizeIngressBuff)
	d.Buffers.Egress = make(chan *shila.Packet, d.config.SizeEgressBuff)

	return nil
}

func (d *Device) TearDown() error {

	if !d.IsSetup() {
		return nil
	}

	// Stop the reader and the writer
	err := d.Stop()

	// Remove the routing table associated with the kernel endpoint
	err = d.removeRouting()

	// Deallocate the corresponding instance of the interface
	err = d.vif.TurnDown()
	err = d.vif.Teardown()

	d.vif = nil
	d.Buffers.Ingress = nil
	d.Buffers.Egress = nil

	return err
}

func (d *Device) Start() error {

	if !d.IsSetup() {
		return Error(fmt.Sprint("Cannot start kernel endpoint ",
			d.Id.Name(), " - ", "Device not yet setup."))
	}

	if d.IsRunning() {
		return Error(fmt.Sprint("Cannot start kernel endpoint ",
			d.Id.Name(), " - ", "Device already running."))

	}

	d.isRunning = true
	return nil
}

func (d *Device) Stop() error {

	if !d.IsSetup() {
		return Error(fmt.Sprint("Cannot stop kernel endpoint ",
			d.Id.Name(), " - ", "Device not yet setup."))
	}

	if !d.IsRunning() {
		return Error(fmt.Sprint("Cannot stop kernel endpoint ",
			d.Id.Name(), " - ", "Device is not running."))

	}

	d.isRunning = false
	return nil
}

func (d *Device) IsSetup() bool {
	return d.vif != nil && d.Buffers.Ingress != nil && d.Buffers.Egress != nil
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
