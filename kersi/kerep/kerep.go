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
	Id         Identifier
	config     config.KernelEndpoint
	ingressBuf chan *shila.Packet
	egressBuf  chan *shila.Packet
	vif        *vif.Device
	isRunning  bool
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func New(id Identifier, config config.KernelEndpoint,
	ingressBuff chan *shila.Packet, egressBuff chan *shila.Packet) *Device {
	return &Device{id, config, ingressBuff, egressBuff, nil, false}
}

func (d *Device) Setup() error {

	if d.IsSetup() {
		return Error(fmt.Sprint("Unable to setup kernel endpoint ",
			d.Id.Name(), " - ", "Device already setup."))
	}

	// Allocate the vif
	d.vif = vif.New(d.Id.Name(), d.Id.namespace, d.Id.Subnet())

	// Setup the vif
	if err := d.vif.Setup(); err != nil {
		d.vif = nil
		return err
	}

	// Turn the vif up
	if err := d.vif.TurnUp(); err != nil {
		_ = d.vif.Teardown()
		d.vif = nil
		return nil
	}

	// Setup the routing
	if err := d.setupRouting(); err != nil {
		_ = d.vif.TurnDown()
		_ = d.vif.Teardown()
		d.vif = nil
		return err
	}

	// Allocate the buffers
	d.ingressBuf = make(chan *shila.Packet, d.config.SizeIngressBuff)
	d.egressBuf = make(chan *shila.Packet, d.config.SizeEgressBuff)

	return nil
}

func (d *Device) TearDown() error {

	if !d.IsSetup() {
		return Error(fmt.Sprint("Unable to tear down kernel endpoint ",
			d.Id.Name(), " - ", "Device not even setup."))
	}

	var err error
	// Return the most recent error, if there is one.
	// However we proceed with the teardown nevertheless.

	// Stop the reader and the writer
	err = d.Stop()

	// Remove the routing table associated with the kernel endpoint
	err = d.removeRouting()

	// Deallocate the corresponding instance of the interface
	err = d.vif.TurnDown()
	err = d.vif.Teardown()

	d.vif = nil
	d.ingressBuf = nil
	d.egressBuf = nil

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
	return d.vif != nil && d.ingressBuf != nil && d.egressBuf != nil
}

func (d *Device) IsRunning() bool {
	return d.isRunning
}

func (d *Device) setupRouting() error {

	// ip rule add from <dev ip> table <table id>
	args := []string{"rule", "add", "from", d.Id.subnet.IP.String(), "table", fmt.Sprint(d.Id.number)}
	if err := helper.ExecuteIpCommand(d.Id.namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", err.Error()))
	}

	// ip route add table <table id> default dev <dev name> scope link
	args = []string{"route", "add", "table", fmt.Sprint(d.Id.number), "default", "dev", d.Id.Name(), "scope", "link"}
	if err := helper.ExecuteIpCommand(d.Id.namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to setup routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", err.Error()))
	}

	return nil
}

func (d *Device) removeRouting() error {

	// ip rule del table <table id>
	args := []string{"rule", "del", "table", fmt.Sprint(d.Id.number)}
	if err := helper.ExecuteIpCommand(d.Id.namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to remove routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", err.Error()))
	}

	// ip route flush table <table id>
	args = []string{"route", "flush", "table", fmt.Sprint(d.Id.number)}
	if err := helper.ExecuteIpCommand(d.Id.namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to remove routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", err.Error()))
	}

	return nil
}
