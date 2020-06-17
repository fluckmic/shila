package router

import (
	"shila/core/shila"
	"shila/networkSide/network"
)

type Response struct {
	Dst    shila.NetworkAddress
	Path   network.Path
	From   ResponseLabel
	IPFlow shila.IPFlow
}

type ResponseLabel uint8
const (
	_                       = iota
	IPOptions ResponseLabel = iota
	MPTCPEndpointToken
	RoutingTable
	Unknown
)

func(label ResponseLabel) String() string {
	switch label {
	case IPOptions: 			return "IPOptions"
	case MPTCPEndpointToken:  	return "MPTCPEndpointToken"
	case RoutingTable:			return "RoutingTable"
	}
	return "Unknown"
}
