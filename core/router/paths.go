package router

import (
	"fmt"
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/config"
	"shila/core/shila"
	"sort"
)

type paths struct {
	storage 	[]PathWrapper
	mapping 	map[shila.IPFlowKey] int
}

type PathWrapper struct {
	path  		snet.Path
	nUsed 		int
	rawMetric 	int
}

// If there is any error in the creation of the paths we just do not specify any. This is oke.
func newPaths(dstAddr shila.NetworkAddress) paths {

	scionPaths := fetchSCIONPaths(dstAddr)
	if scionPaths == nil {
		return paths{}
	}

	sortPaths(scionPaths)

	return paths{
		storage: scionPaths,
		mapping: make(map[shila.IPFlowKey] int),
	}
}

func (p *paths) get(key shila.IPFlowKey) (*PathWrapper, int) {

	if (p.storage == nil) || (len(p.storage) < 1) {
		return nil, -1
	}
	
	useCountOfNext := len(p.mapping) / len(p.storage) // #used / #total

	for index, pathWrapper := range p.storage {
		if pathWrapper.nUsed == useCountOfNext {
			p.storage[index].nUsed++
			p.mapping[key] = index

			return &pathWrapper, len(p.mapping)
		}
	}

	return nil, -1
}

func (p *paths) free(key shila.IPFlowKey) {
	if index, ok := p.mapping[key]; ok {
		p.storage[index].nUsed--
		delete(p.mapping, key)
	}
}

func fetchSCIONPaths(dstAddr shila.NetworkAddress) []PathWrapper {
	dstAddrIA := dstAddr.(*snet.UDPAddr).IA
	if paths, err := appnet.QueryPaths(dstAddrIA); err != nil {
		return nil
	} else {
		pathsWrapped := make([]PathWrapper, 0, len(paths))
		for i, path := range paths {
			fmt.Printf("[%2d] %s\n", i, fmt.Sprintf("%s", path))
			switch selectPathAlgorithm() {
			case appnet.Shortest:
				pathsWrapped = append(pathsWrapped, PathWrapper{path: path, nUsed: 0, rawMetric: len(path.Interfaces())})
			case appnet.MTU:
				pathsWrapped = append(pathsWrapped, PathWrapper{path: path, nUsed: 0, rawMetric: int(path.MTU())})
			}
		}
		return pathsWrapped
	}
}

func sortPaths(paths []PathWrapper) {
	switch selectPathAlgorithm() {
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

func selectPathAlgorithm() int {
	if config.Config.Router.PathSelection == "mtu" {
		return appnet.MTU
	} else if config.Config.Router.PathSelection == "shortest" {
		return appnet.Shortest
	} else {
		return appnet.Shortest
	}
}
