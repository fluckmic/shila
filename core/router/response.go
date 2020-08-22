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
	MainTCPFlow  shila.TCPFlow
	FlowCount    int
	RawMetrics   []int
	Sharability  int
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