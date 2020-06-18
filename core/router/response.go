//
package router

import (
	"github.com/bclicn/color"
	"shila/core/shila"
)

type Entry struct {
	Dst   shila.NetworkAddress
	Paths paths
}

type Response struct {
	Dst    			shila.NetworkAddress
	Path   			shila.NetworkPath
	FlowCategory	FlowCategory
	MainIPFlowKey 	shila.IPFlowKey
}

type FlowCategory uint8
const (
	_                      = iota
	MainFlow FlowCategory = iota
	SubFlow
	Unknown
)

func(label FlowCategory) String() string {
	switch label {
	case MainFlow: 	return color.LightBlue("MainFlow")
	case SubFlow:  	return color.LightPurple("SubFlow")
	}
	return "Unknown"
}