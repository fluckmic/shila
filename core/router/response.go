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
	Dst          shila.NetworkAddress
	Path         shila.NetworkPath
	FlowCategory FlowCategory
	MainIPFlow   shila.IPFlow
	FlowCount    int
	RawMetrics   []int
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