package connection

import "github.com/bclicn/color"

type Category uint8
const (
	_                 = iota
	MainFlow Category = iota
	SubFlow
	Unknown
)

func(cat Category) String() string {
	switch cat {
	case MainFlow: 	return color.LightBlue("MainFlow")
	case SubFlow:  	return color.LightPurple("SubFlow")
	}
	return "Unknown"
}
