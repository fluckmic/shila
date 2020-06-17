//
package router

import (
	"shila/core/shila"
	"shila/networkSide/network"
)

type Entry struct {
	Dst	 	shila.NetworkAddress
	Paths 	[]shila.NetworkPath
}

type Response struct {
	Dst    			shila.NetworkAddress
	Path   			network.Path
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
	case MainFlow: 	return "MainFlow"
	case SubFlow:  	return "SubFlow"
	}
	return "Unknown"
}