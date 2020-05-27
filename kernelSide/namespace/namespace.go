package namespace

import (
	"os/exec"
)

type Namespace struct {
	Name     string
	NonEmpty bool
}

func NewNamespace(name string) Namespace {
	return Namespace{
		Name:  		name,
		NonEmpty:	true,
	}
}

// ip netns add <namespace name>
func AddNamespace(namespace Namespace) error {
	if err := execIpCmd("netns", "add", namespace.Name); err != nil {
		return err
	}
	return nil
}

// ip netns delete <namespace name>
func DeleteNamespace(namespace Namespace) error {
	if err := execIpCmd("netns", "delete", namespace.Name); err != nil {
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

func execIpCmdInNamespace(namespace Namespace, args ...string) error {
	return execIpCmd(append([]string{"netns", "exec", namespace.Name, "ip"}, args...)...)
}

func Execute(namespace Namespace, args ...string) error {
	if namespace.NonEmpty {
		return execIpCmdInNamespace(namespace, args...)
	} else {
		return execIpCmd(args...)
	}
}