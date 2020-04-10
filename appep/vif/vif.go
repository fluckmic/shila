// TODO: vif package deserves a description.
package vif

import (
	"fmt"
	"shila/appep/tun"
)

type Error string
func (e Error) Error() string {
	return string(e)
}

type Device struct {
	Name 	  string
	Namespace Namespace
	Address   string
	device 	  *tun.Device
}

type Namespace struct {
	Name string
}

func New(name string, namespace Namespace, address string) *Device {
	return &Device{name, namespace, address, nil}
}

// TODO: Setup deserves a description.
func (d *Device) Setup() error {

	if d.device != nil {
		return Error(fmt.Sprint("Unable to setup vif device: ",
								d.Name, " - ", "Device already setup."))
	}

	// Create and allocated a new tun device
	d.device = tun.New(d.Name)
	if err := d.device.Allocate(); err != nil {
		return err
	}

	// TODO: Put the tun device into the correct namespace
	/*
	cmd := exec.Command("ip", "link", "set", d.Name, "netns", d.Namespace.Name)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(out.String())
	*/

	// TODO: Assign the network

	return nil
}

// TODO: Up deserves a description.
func (d *Device) Up() error {
	return nil
}

// TODO: Down deserves a description.
func (d *Device) Down() error {
	return nil
}

// TODO: Teardown deserves a description
func (d *Device) Teardown() error {
	return nil
}