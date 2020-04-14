package helper

import "os/exec"

type Error string

func (e Error) Error() string {
	return string(e)
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
