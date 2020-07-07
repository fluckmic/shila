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

func getLengthOptSubset(paths []PathWrapper) (optSubset []PathWrapper, sharabilityValue int) {
	sortPaths(paths, length)
	optSubset = trowAwayOddPaths(paths)
	sharabilityValue = calculateSharabilityForPaths(optSubset)
	return
}

func getMtuOptSubset(paths []PathWrapper) (optSubset []PathWrapper, sharabilityValue int) {
	sortPaths(paths, mtu)
	optSubset = trowAwayOddPaths(paths)
	sharabilityValue = calculateSharabilityForPaths(optSubset)
	return
}

func trowAwayOddPaths(paths []PathWrapper) []PathWrapper {
	if len(paths) <= config.Config.KernelSide.NumberOfEgressInterfaces {
		return paths
	} else {
		subset := make([]PathWrapper, config.Config.KernelSide.NumberOfEgressInterfaces)
		copy(subset, paths[0:config.Config.KernelSide.NumberOfEgressInterfaces])
		return subset
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
	} else if config.Config.Router.PathSelection == "length" {
		return length
	} else if config.Config.Router.PathSelection == "sharability" {
		return sharability
	} else {
		log.Error.Println("Unknown path selection; using", mtu, ".")
		return mtu
	}
}