package testing

import (
	"shila/appep"
	"testing"
)

func TestDevice_Setup(t *testing.T) {

	var appEp appep.Device

	if err := appEp.Setup(); err == nil {
		t.Error("Cannot setup a device before initialization.")
	}

}
