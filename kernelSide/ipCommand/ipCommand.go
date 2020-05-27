package ipCommand

import (
	"os/exec"
)

type Namespace struct {
	Name string
}

// ip netns add <namespace name>
func AddNamespace(name string) error {
	if err := execIpCmd("netns", "add", name); err != nil {
		return err
	}
	return nil
}

// ip netns delete <namespace name>
func DeleteNamespace(name string) error {
	if err := execIpCmd("netns", "delete", name); err != nil {
		return err
	}
	return nil
}

func execIpCmd(args ...string) error {

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

func execIpCmdInNamespace(ns string, args ...string) error {
	return execIpCmd(append([]string{"netns", "exec", ns, "ip"}, args...)...)
}

func Execute(ns *Namespace, args ...string) error {
	if ns == nil {
		return execIpCmd(args...)
	} else {
		return execIpCmdInNamespace(ns.Name, args...)
	}
}
