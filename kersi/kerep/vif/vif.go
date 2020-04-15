// TODO: vif package deserves a description.
package vif

import (
	"fmt"
	"shila/helper"
	"shila/kersi/kerep/tun"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type Device struct {
	Name      string
	Namespace *helper.Namespace
	Subnet    string
	device    *tun.Device
	isUp      bool
}

func New(name string, namespace *helper.Namespace, subnet string) *Device {
	return &Device{name, namespace, subnet, nil, false}
}

// Setup allocates the vif device.
func (d *Device) Setup() error {

	if d.IsSetup() {
		return Error(fmt.Sprint("Unable to setup vif device ",
			d.Name, " - ", "Device already setup."))
	}

	// Create and allocated a new tun device
	d.device = tun.New(d.Name)
	if err := d.device.Allocate(); err != nil {
		return err
	}

	// Assign the device to the namespace
	if err := d.assignNamespace(); err != nil {
		return err
	}

	// Assign subnet to the device
	if err := d.assignSubnet(); err != nil {
		return err
	}

	return nil
}

// Teardown de-allocate and deletes the vif device.
func (d *Device) Teardown() error {

	if !d.IsSetup() {
		return Error(fmt.Sprint("Unable to tear down vif device ", d.Name,
			" - ", "Device not even setup."))
	}

	var err error
	// Return the most recent error, if there is one.
	// However we proceed with the teardown nevertheless.

	// Remove the interface from the system
	err = d.removeInterface()

	// Deallocate the corresponding instance of the interface
	err = d.device.Deallocate()

	d.device = nil

	return err
}

// TurnUp enables the vif device.
func (d *Device) TurnUp() error {

	if !d.IsSetup() {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name,
			" up - ", "Device not even setup."))
	}

	if d.IsUp() {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name,
			" up - ", "Device already up."))
	}

	args := []string{"link", "set", d.Name, "up"}
	if err := helper.ExecuteIpCommand(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name, " up - ", err.Error()))
	}
	d.isUp = true
	return nil
}

// TurnDown disables the vif device.
func (d *Device) TurnDown() error {

	if !d.IsSetup() {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name,
			" down - ", "Device not even setup."))
	}

	if !d.IsUp() {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name,
			" down - ", "Device already down."))
	}

	// ip link set <device name> down
	args := []string{"link", "set", d.Name, "down"}
	if err := helper.ExecuteIpCommand(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name, " down - ", err.Error()))
	}
	d.isUp = false
	return nil
}

func (d *Device) Read(b []byte) (int, error) {

	if !d.IsSetup() {
		err := Error(fmt.Sprint("Cannot read from vif device ", d.Name,
			" - ", "Device not even setup."))
		return 0, err
	}

	if !d.IsUp() {
		err := Error(fmt.Sprint("Cannot read from vif device ", d.Name,
			" - ", "Device is not up."))
		return 0, err
	}

	return d.device.Read(b)
}

func (d *Device) Write(b []byte) (int, error) {

	if !d.IsSetup() {
		err := Error(fmt.Sprint("Cannot write to vif device ", d.Name,
			" - ", "Device not even setup."))
		return 0, err
	}

	if !d.IsUp() {
		err := Error(fmt.Sprint("Cannot write to vif device ", d.Name,
			" - ", "Device is not up."))
		return 0, err
	}

	return d.device.Write(b)
}

func (d *Device) IsSetup() bool {
	return d.device != nil
}

func (d *Device) IsUp() bool {
	return d.isUp
}

func (d *Device) assignNamespace() error {

	// Nothing to do if there is no namespace to be assigned
	if d.Namespace == nil {
		return nil
	}

	// ip link set <device name> netns <namespace name>
	err := helper.ExecuteIpCommand(nil, "link", "set", d.Name, "netns", d.Namespace.Name)
	if err != nil {
		return Error(fmt.Sprint("Unable to assign namespace ", d.Namespace.Name,
			" to vif device ", d.Name, " - ", err.Error()))
	}

	return nil
}

// assignSubnet assign the subnet to the vif device. If the vif device is part of a namespace
// then it is assumed that the device is already part of this namespace, i.e. that there was
// was already a successful call to assignNamespace().
func (d *Device) assignSubnet() error {

	// ip addr add <subnet> dev <dev name>
	args := []string{"addr", "add", d.Subnet, "dev", d.Name}
	if err := helper.ExecuteIpCommand(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to assign subnet ", d.Subnet,
			" to vif device ", d.Name, " - ", err.Error()))
	}
	return nil
}

func (d *Device) removeInterface() error {

	// ip link delete <interface name>
	args := []string{"link", "delete", d.Name}
	if err := helper.ExecuteIpCommand(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to remove interface ", d.Name, " - ", err.Error()))
	}
	return nil
}
