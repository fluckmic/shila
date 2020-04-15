package helper

import "os/exec"

type Error string

func (e Error) Error() string {
	return string(e)
}

func AddNamespace(name string) error {
	// ip netns add <namespace name>
	if err := ExecuteIpCommand("netns", "add", name); err != nil {
		return err
	}
	return nil
}

func DeleteNamespace(name string) error {
	// ip netns delete <namespace name>
	if err := ExecuteIpCommand("netns", "delete", name); err != nil {
		return err
	}
	return nil
}

func ExecuteIpCommand(args ...string) error {

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
