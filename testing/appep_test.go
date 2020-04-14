package testing

import (
	"shila/kersi/kerep"
	"testing"
)

func TestDevice_Setup(t *testing.T) {

	var kernelEndpoint kerep.Device

	if err := kernelEndpoint.Setup(); err == nil {
		t.Error("Cannot setup a device before initialization.")
	}

}
