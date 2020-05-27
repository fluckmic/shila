
package vif

import (
	"fmt"
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint/tun"
	"shila/kernelSide/namespace"
)

type Device struct {
	Name      	string
	Namespace 	namespace.Namespace
	Subnet    	string
	device    	tun.Device
	state   	shila.EntityState
}

func New(name string, namespace namespace.Namespace, subnet string) *Device {
	return &Device{
		Name: 		name,
		Namespace: 	namespace,
		Subnet: 	subnet,
		state:		shila.NewEntityState(),
	}
}

// Setup allocates the vif device.
func (d *Device) Setup() error {

	if d.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}

	// Create and allocated a new tun device
	d.device = tun.New(d.Name)
	if err := d.device.Allocate(); err != nil {
		_ = d.device.Deallocate()
		return err
	}

	// Assign the device to the namespace
	if err := d.assignNamespace(); err != nil {
		_ = d.device.Deallocate()
		return err
	}

	// Assign subnet to the device
	if err := d.assignSubnet(); err != nil {
		_ = d.removeInterface()
		_ = d.device.Deallocate()
		return err
	}

	d.state.Set(shila.Initialized)
	return nil
}

// Teardown de-allocate and deletes the vif device.
func (d *Device) Teardown() error {
	var err error
	err = d.removeInterface()
	err = d.device.Deallocate()
	d.state.Set(shila.TornDown)
	return err
}

// TurnUp enables the vif device.
func (d *Device) TurnUp() error {

	if d.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}

	args := []string{"link", "set", d.Name, "up"}
	if err := namespace.Execute(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name, " up - ", err.Error()))
	}

	d.state.Set(shila.Running)
	return nil
}

// TurnDown disables the vif device.
func (d *Device) TurnDown() error {

	// ip link set <device name> down
	args := []string{"link", "set", d.Name, "down"}
	if err := namespace.Execute(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to turn vif device ", d.Name, " down - ", err.Error()))
	}

	d.state.Set(shila.Initialized)
	return nil
}

func (d *Device) Read(b []byte) (int, error) {
	if d.state.Not(shila.Running) {
		return -1, shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}
	return d.device.Read(b)
}

func (d *Device) Write(b []byte) (int, error) {
	if d.state.Not(shila.Running) {
		return -1, shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}
	return d.device.Write(b)
}

func (d *Device) assignNamespace() error {

	// Nothing to do if there is no namespace to be assigned
	if !d.Namespace.NonEmpty {
		return nil
	}

	// ip link set <device name> netns <namespace name>
	err := namespace.Execute(namespace.Namespace{}, "link", "set", d.Name, "netns", d.Namespace.Name)
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
	if err := namespace.Execute(d.Namespace, args...); err != nil {
		return Error(fmt.Sprint("Unable to assign subnet ", d.Subnet,
			" to vif device ", d.Name, " - ", err.Error()))
	}
	return nil
}

func (d *Device) removeInterface() error {

	// ip link delete <interface name>
	args := []string{"link", "delete", d.Name}
	err := namespace.Execute(d.Namespace, args...)
	return err
}
