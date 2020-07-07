package router

import (
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
)

type paths struct {
	storage 	[]PathWrapper
	mapping 	map[shila.IPFlowKey] int
	sharability int
}

type PathWrapper struct {
	path  		snet.Path
	edgeIndices []int
	nUsed 		int
	rawMetrics 	[]int
}

// If there is any error in the creation of the paths we just do not specify any. This is oke.
func newPaths(dstAddr shila.NetworkAddress) paths {

	scionPaths := fetchAndWrapSCIONPaths(dstAddr)
	if scionPaths == nil {
		return paths{}
	}

	sharabilityValue := -1
	switch selectPathAlgorithm() {
	case mtu:
		scionPaths, sharabilityValue = getMtuOptSubset(scionPaths)

	case length:
		scionPaths, sharabilityValue = getLengthOptSubset(scionPaths)

	case sharability:
		scionPaths, sharabilityValue = getSharabilityOptSubset(scionPaths)
	default:
		return paths{}
	}

	return paths{
		storage: 		scionPaths,
		mapping: 		make(map[shila.IPFlowKey] int),
		sharability: 	sharabilityValue,
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

func fetchAndWrapSCIONPaths(dstAddr shila.NetworkAddress) []PathWrapper {
	dstAddrIA := dstAddr.(*snet.UDPAddr).IA
	if paths, err := appnet.QueryPaths(dstAddrIA); err != nil {
		return nil
	} else {
		pathsWrapped := make([]PathWrapper, 0, len(paths))
		for _, path := range paths {
			rawMetrics := []int{int(path.MTU()), len(path.Interfaces())}
			pathsWrapped = append(pathsWrapped, PathWrapper{path: path, nUsed: 0, rawMetrics: rawMetrics })
			//fmt.Printf("[%2d] %s\n", i, fmt.Sprintf("%s", path))
		}
		return pathsWrapped
	}
}