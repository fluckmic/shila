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
func (device *Device) Setup() error {

	if device.state.Not(shila.Uninitialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}

	// Create and allocated a new tun device
	device.device = tun.New(device.Name)
	if err := device.device.Allocate(); err != nil {
		_ = device.device.Deallocate()
		return shila.PrependError(err, fmt.Sprint("Unable to allocate device ", device.Name, "."))
	}

	// Assign the device to the namespace
	if err := device.assignNamespace(); err != nil {
		_ = device.device.Deallocate()
		return shila.PrependError(err, fmt.Sprint("Unable to assign namespace."))
	}

	// Assign subnet to the device
	if err := device.assignSubnet(); err != nil {
		_ = device.removeInterface()
		_ = device.device.Deallocate()
		return shila.PrependError(err, fmt.Sprint("Unable to assign subnet."))
	}

	device.state.Set(shila.Initialized)
	return nil
}

// Teardown de-allocate and deletes the vif device.
func (device *Device) Teardown() error {
	device.state.Set(shila.TornDown)
	var err error
	err = device.removeInterface()
	err = device.device.Deallocate()
	return err
}

// TurnUp enables the vif device.
func (device *Device) TurnUp() error {

	if device.state.Not(shila.Initialized) {
		return shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}

	args := []string{"link", "set", device.Name, "up"}
	if err := network.Execute(device.Namespace, args...); err != nil {
		return err
	}

	device.state.Set(shila.Running)
	return nil
}

// TurnDown disables the vif device.
func (device *Device) TurnDown() error {

	// ip link set <device name> down
	args := []string{"link", "set", device.Name, "down"}
	if err := network.Execute(device.Namespace, args...); err != nil {
		return err
	}

	device.state.Set(shila.Initialized)
	return nil
}

func (device *Device) Read(b []byte) (int, error) {
	if device.state.Not(shila.Running) {
		return -1, shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}
	return device.device.Read(b)
}

func (device *Device) Write(b []byte) (int, error) {
	if device.state.Not(shila.Running) {
		return -1, shila.CriticalError(fmt.Sprint("Entity in wrong state ", device.state, "."))
	}
	return device.device.Write(b)
}

func (device *Device) assignNamespace() error {

	// Nothing to do if there is no namespace to be assigned
	if !device.Namespace.NonEmpty {
		return nil
	}

	// ip link set <device name> netns <namespace name>
	err := network.Execute(network.Namespace{}, "link", "set", device.Name, "netns", device.Namespace.Name)
	if err != nil {
		return err
	}

	return nil
}

// assignSubnet assign the subnet to the vif device. If the vif device is part of a namespace
// then it is assumed that the device is already part of this namespace, i.e. that there was
// was already a successful call to assignNamespace().
func (device *Device) assignSubnet() error {
	// ip addr add <subnet> dev <dev name>
	args := []string{"addr", "add", string(device.Subnet), "dev", device.Name}
	if err := network.Execute(device.Namespace, args...); err != nil {
		return err
	}
	return nil
}

func (device *Device) removeInterface() error {
	// ip link delete <interface name>
	args := []string{"link", "delete", device.Name}
	err := network.Execute(device.Namespace, args...)
	return err
}

func (device *Device) Says(str string) string {
	return  fmt.Sprint(device.Identifier(), ": ", str)
}

func (device *Device) Identifier() string {
	return device.Name
}