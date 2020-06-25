package router

import (
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
	"sort"
)

const (
	pathSelectionConfig = appnet.Shortest
)

type paths struct {
	storage 	[]pathWrapper
	mapping 	map[shila.IPFlowKey] int
}

type pathWrapper struct {
	path  snet.Path
	nUsed int
}

// If there is any error in the creation of the paths we just do not specify any. This is oke.
func newPaths(dstAddr shila.NetworkAddress) paths {

	scionPaths := fetchSCIONPaths(dstAddr)
	if scionPaths == nil {
		return paths{}
	}

	sortPaths(scionPaths, pathSelectionConfig)

	return paths{
		storage: scionPaths,
		mapping: make(map[shila.IPFlowKey] int),
	}
}

func (p *paths) get(key shila.IPFlowKey) shila.NetworkPath {

	if (p.storage == nil) || (len(p.storage) < 1) {
		return nil
	}
	
	useCountOfNext := len(p.mapping) / len(p.storage) // #used / #total

	for index, pathWrapper := range p.storage {
		if pathWrapper.nUsed == useCountOfNext {
			p.storage[index].nUsed++
			p.mapping[key] = index

			return pathWrapper.path
		}
	}

	return nil
}

func (p *paths) free(key shila.IPFlowKey) {
	if index, ok := p.mapping[key]; ok {
		p.storage[index].nUsed--
		delete(p.mapping, key)
	}
}

func fetchSCIONPaths(dstAddr shila.NetworkAddress) []pathWrapper {
	dstAddrIA 	:= dstAddr.(*snet.UDPAddr).IA
	if paths, err := appnet.QueryPaths(dstAddrIA); err != nil {
		return nil
	} else {
		pathsWrapped := make([]pathWrapper, 0, len(paths))
		for _, path := range paths {
			pathsWrapped = append(pathsWrapped, pathWrapper{path: path, nUsed: 0})
		}
		return pathsWrapped
	}
}

func sortPaths(paths []pathWrapper, pathAlgorithm int) {

	switch pathAlgorithm {
	case appnet.Shortest:
		sort.Slice(paths, func(i, j int) bool {
			return len(paths[i].path.Interfaces()) < len(paths[j].path.Interfaces())
		})
	case appnet.MTU: {
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].path.MTU() > paths[j].path.MTU()
		})
	}
	default:
		return
	}
	return
}
