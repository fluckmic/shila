//
package vif

import (
	"fmt"
	"shila/core/shila"
	"shila/kernelSide/kernelEndpoint/tun"
	"shila/kernelSide/network"
)

type Device struct {
	Name      string
	Namespace network.Namespace
	Subnet    network.Subnet
	device    tun.Device
	state     shila.EntityState
}

func New(name string, namespace network.Namespace, subnet network.Subnet) Device {
	return Device{
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
		return shila.PrependError(err, fmt.Sprint("Unable to allocate device {", d.Name, "}."))
	}

	// Assign the device to the namespace
	if err := d.assignNamespace(); err != nil {
		_ = d.device.Deallocate()
		return shila.PrependError(err, fmt.Sprint("Unable to assign namespace."))
	}

	// Assign subnet to the device
	if err := d.assignSubnet(); err != nil {
		_ = d.removeInterface()
		_ = d.device.Deallocate()
		return shila.PrependError(err, fmt.Sprint("Unable to assign subnet."))
	}

	d.state.Set(shila.Initialized)
	return nil
}

// Teardown de-allocate and deletes the vif device.
func (d *Device) Teardown() error {
	d.state.Set(shila.TornDown)
	var err error
	err = d.removeInterface()
	err = d.device.Deallocate()
	return err
}

// TurnUp enables the vif device.
func (d *Device) TurnUp() error {

	if d.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state {", d.state, "}."))
	}

	args := []string{"link", "set", d.Name, "up"}
	if err := network.Execute(d.Namespace, args...); err != nil {
		return err
	}

	d.state.Set(shila.Running)
	return nil
}

// TurnDown disables the vif device.
func (d *Device) TurnDown() error {

	// ip link set <device name> down
	args := []string{"link", "set", d.Name, "down"}
	if err := network.Execute(d.Namespace, args...); err != nil {
		return err
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
	err := network.Execute(network.Namespace{}, "link", "set", d.Name, "netns", d.Namespace.Name)
	if err != nil {
		return err
	}

	return nil
}

// assignSubnet assign the subnet to the vif device. If the vif device is part of a namespace
// then it is assumed that the device is already part of this namespace, i.e. that there was
// was already a successful call to assignNamespace().
func (d *Device) assignSubnet() error {
	// ip addr add <subnet> dev <dev name>
	args := []string{"addr", "add", string(d.Subnet), "dev", d.Name}
	if err := network.Execute(d.Namespace, args...); err != nil {
		return err
	}
	return nil
}

func (d *Device) removeInterface() error {
	// ip link delete <interface name>
	args := []string{"link", "delete", d.Name}
	err := network.Execute(d.Namespace, args...)
	return err
}
