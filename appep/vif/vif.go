// TODO: vif package deserves a description.
package vif

import (
	"fmt"
	"os/exec"
	"shila/appep/tun"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

type Device struct {
	Name      string
	Namespace *Namespace
	Subnet    string
	device    *tun.Device
}

type Namespace struct {
	Name string
}

func New(name string, namespace *Namespace, address string) *Device {
	return &Device{name, namespace, address, nil}
}

// TODO: Setup deserves a description.
func (d *Device) Setup() error {

	if d.device != nil {
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

// TODO: Teardown deserves a description.
func (d *Device) Teardown() error {
	return nil
}

// TODO: TurnUp deserves a description.
func (d *Device) TurnUp() error {

	var args []string
	if d.Namespace == nil {
		args = []string{"link", "set", d.Name, "up"}
	} else {
		args = []string{"netns", "exec", d.Namespace.Name, "ip", "link", "set", d.Name, "up"}
	}

	err := executeIpCommand(args...)
	if err != nil {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name, "up - ", err.Error()))
	}

	return nil
}

// TODO: TurnDown deserves a description.
func (d *Device) TurnDown() error {
	return nil
}

func (d *Device) assignNamespace() error {

	// Nothing to do if there is no namespace to be assigned
	if d.Namespace == nil {
		return nil
	}

	err := executeIpCommand("link", "set", d.Name, "netns", d.Namespace.Name)
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

	var args []string
	if d.Namespace == nil {
		args = []string{"addr", "add", d.Subnet, "dev", d.Name}
	} else {
		args = []string{"netns", "exec", d.Namespace.Name, "ip", "addr", "add", d.Subnet, "dev", d.Name}
	}

	err := executeIpCommand(args...)
	if err != nil {
		return Error(fmt.Sprint("Unable to assign subnet ", d.Subnet,
			" to vif device ", d.Name, " - ", err.Error()))
	}

	return nil
}

func executeIpCommand(args ...string) error {

	_, err := exec.Command("ip", args...).Output()
	if err != nil {
		// From the official documentation: "Any returned error will *usually be of type *ExitError."
		if exitError, ok := err.(*exec.ExitError); ok {
			return Error(exitError.Stderr)
		} else {
			return err
		}
	}

	return nil
}
