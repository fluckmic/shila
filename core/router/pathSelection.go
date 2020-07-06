package router

import (
	"shila/config"
	"shila/log"
	"sort"
)

const (
	nPathSelections = 3
)

type pathSelection uint8
const (
	mtu pathSelection = 0
	length			  = 1
	sharability		  = 2
)

func(ps pathSelection) String() string {
	switch ps {
	case mtu: 			return "mtu"
	case length: 		return "length"
	case sharability:	return "sharability"
	}
	return "Unknown"
}

func getLengthOptSubset(scionPaths []PathWrapper) []PathWrapper {
	sortPaths(scionPaths, length)
	return truncatePaths(scionPaths)	// Not necessary to carry around more paths than needed.
}

func getMtuOptSubset(scionPaths []PathWrapper) []PathWrapper {
	sortPaths(scionPaths, mtu)
	return truncatePaths(scionPaths)
}

func truncatePaths(scionPaths []PathWrapper) []PathWrapper {
	if len(scionPaths) <= config.Config.KernelSide.NumberOfEgressInterfaces {
		return scionPaths
	} else {
		truncScionPaths := make([]PathWrapper, config.Config.KernelSide.NumberOfEgressInterfaces)
		copy(truncScionPaths, scionPaths[0:config.Config.KernelSide.NumberOfEgressInterfaces])
		return truncScionPaths
	}
}

func sortPaths(paths []PathWrapper, selection pathSelection) {
	switch selection {
	case length:
		sort.Slice(paths, func(i, j int) bool {
			return len(paths[i].path.Interfaces()) < len(paths[j].path.Interfaces())
		})
	case mtu: {
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].path.MTU() > paths[j].path.MTU()
		})
	}
	default:
		return
	}
	return
}

func selectPathAlgorithm() pathSelection {
	if config.Config.Router.PathSelection == "mtu" {
		return mtu
	} else if config.Config.Router.PathSelection == "shortest" {
		return length
	} else if config.Config.Router.PathSelection == "sharability" {
		return sharability
	} else {
		log.Error.Println("Unknown path selection; using", mtu, ".")
		return mtu
	}
}