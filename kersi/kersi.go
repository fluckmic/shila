package kersi

import (
	"fmt"
	"shila/config"
)

type Manager struct {
	config config.Config
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func New(config config.Config) *Manager {
	return &Manager{config}
}

func (m *Manager) Setup() error {
	return nil
}

func (m *Manager) CleanUp() {
	fmt.Print("Called CleanUp.")
	// Clean up the kernel side as good as possible.
}

func (m *Manager) Start() error {

	if !m.IsSetup() {
		return Error(fmt.Sprint("Cannot start kernel side",
			" - ", "Kernel side not yet setup."))
	}

	if m.IsRunning() {
		return Error(fmt.Sprint("Cannot start kernel side",
			" - ", "Kernel side already running."))

	}

	return nil
}

func (m *Manager) Stop() error {

	if !m.IsSetup() {
		return Error(fmt.Sprint("Cannot stop kernel side",
			" - ", "Device not yet setup."))
	}

	if !m.IsRunning() {
		return Error(fmt.Sprint("Cannot stop kernel side",
			" - ", "Device is not running."))

	}

	return nil
}

func (m *Manager) IsSetup() bool {
	return false
}

func (m *Manager) IsRunning() bool {
	return false
}

func (m *Manager) setupRouting() error {

	// TODO: Explain what is done here.
	// ip rule add from <dev ip> table <table id>
	// ip route add table <table id> default dev <dev name> scope link
	/*
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
	*/
	return nil
}

func (m *Manager) removeRouting() error {

	// TODO: Explain what is done here.
	// ip rule del table <table id>
	// ip route flush table <table id>
	/*
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
	*/
	return nil
}
