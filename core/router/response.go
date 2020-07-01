//
package router

import (
	"shila/core/shila"
)

type Entry struct {
	Dst   		shila.NetworkAddress
	Paths 		paths
	HumanFlowID	string
}

type Response struct {
	Dst    			shila.NetworkAddress
	Path   			shila.NetworkPath
	FlowCategory	FlowCategory
	MainIPFlow	 	shila.IPFlow
	FlowCount		int
	Quality			int					// For the moment this is the path metric, which is either length or mtu
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