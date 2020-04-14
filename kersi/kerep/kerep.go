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

	// Setup the routing
	if err := d.setupRouting(); err != nil {
		// Return the error which happened in the setup, this is more
		// useful than returning a possible error from the subsequent cleaning up.
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

	if d.IsRunning() {
		return Error(fmt.Sprint("Unable to tear down kernel endpoint ",
			d.Id.Name(), " - ", "Device still running."))
	}

	var err error
	// Return the most recent error, if there is one.
	// However we proceed with the teardown nevertheless.

	// Remove the routing table associated with the kernel endpoint
	err = d.removeRouting()

	// Deallocate the corresponding instance of the interface
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

	return nil
}

func (d *Device) IsSetup() bool {
	return d.vif != nil && d.ingressBuf != nil && d.egressBuf != nil
}

func (d *Device) IsRunning() bool {
	return d.isRunning
}

func (d *Device) setupRouting() error {

	// TODO: Explain what is done here.
	// ip rule add from <dev ip> table <table id>
	// ip route add table <table id> default dev <dev name> scope link

	var argsCmd1, argsCmd2 []string
	if d.Id.InDefaultNamespace() {
		argsCmd1 = []string{"rule", "add", "from", d.Id.subnet.IP.String(), "table", string(d.Id.number)}
		argsCmd2 = []string{"route", "add", "table", string(d.Id.number), "default", "dev", d.Id.Name(), "scope", "link"}
	} else {
		argsCmd1 = []string{"netns", "exec", d.Id.Namespace(), "ip",
			"rule", "add", "from", d.Id.subnet.IP.String(), fmt.Sprint("table", d.Id.number)}
		argsCmd2 = []string{"netns", "exec", d.Id.Namespace(), "ip",
			"route", "add", "table", string(d.Id.number), "default", "dev", d.Id.Name(), "scope", "link"}
	}

	if errCmd1 := helper.ExecuteIpCommand(argsCmd1...); errCmd1 != nil {
		return Error(fmt.Sprint("Unable to setup routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", errCmd1.Error()))
	}

	if errCmd2 := helper.ExecuteIpCommand(argsCmd2...); errCmd2 != nil {
		return Error(fmt.Sprint("Unable to setup routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", errCmd2.Error()))
	}

	return nil
}

func (d *Device) removeRouting() error {

	// TODO: Explain what is done here.
	// ip rule del table <table id>
	// ip route flush table <table id>

	var argsCmd1, argsCmd2 []string
	if d.Id.InDefaultNamespace() {

		argsCmd1 = []string{"rule", "del", "table", string(d.Id.number)}
		argsCmd2 = []string{"route", "flush", "table", string(d.Id.number)}
	} else {
		argsCmd1 = []string{"netns", "exec", d.Id.Namespace(), "ip", "rule", "del", "table", string(d.Id.number)}
		argsCmd2 = []string{"netns", "exec", d.Id.Namespace(), "ip", "route", "flush", "table", string(d.Id.number)}
	}

	if errCmd1 := helper.ExecuteIpCommand(argsCmd1...); errCmd1 != nil {
		return Error(fmt.Sprint("Unable to remove routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", errCmd1.Error()))
	}

	if errCmd2 := helper.ExecuteIpCommand(argsCmd2...); errCmd2 != nil {
		return Error(fmt.Sprint("Unable to remove routing for kernel endpoint ", d.Id.Name(),
			" in namespace ", d.Id.Namespace(), " - ", errCmd2.Error()))
	}

	return nil
}
