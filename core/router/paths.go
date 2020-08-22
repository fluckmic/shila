package router

import (
	"github.com/netsec-ethz/scion-apps/pkg/appnet"
	"github.com/scionproto/scion/go/lib/snet"
	"shila/core/shila"
	"shila/log"
)

type paths struct {
	storage 	[]PathWrapper
	mapping 	map[shila.TCPFlowKey] int
	sharability int
}

type PathWrapper struct {
	path  		snet.Path
	edgeIndices []int
	nUsed 		int
	rawMetrics 	[]int
}

// If there is any error in the creation of the paths we just do not specify any. This is oke.
func newPaths(dstAddr shila.NetworkAddress) (paths, error) {

	scionPaths, err := fetchAndWrapSCIONPaths(dstAddr)
	if err != nil {
		log.Error.Print("Unable to fetch SCION paths. ", err.Error())
		return paths{}, err
	} else if scionPaths == nil {
		// Destination address is in the local IA
		return paths{
			storage: 		[]PathWrapper{{path: nil, rawMetrics: []int{0,0}}},
			mapping: 		make(map[shila.TCPFlowKey] int),
			sharability: 	0,
		}, nil
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
		log.Error.Print("Unknown path selection, using mtu.")
		scionPaths, sharabilityValue = getMtuOptSubset(scionPaths)
	}

	return paths{
		storage: 		scionPaths,
		mapping: 		make(map[shila.TCPFlowKey] int),
		sharability: 	sharabilityValue,
	}, nil
}

func (p *paths) get(key shila.TCPFlowKey) (*PathWrapper, int) {

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

func (p *paths) free(key shila.TCPFlowKey) {
	if index, ok := p.mapping[key]; ok {
		p.storage[index].nUsed--
		delete(p.mapping, key)
	}
}

func fetchAndWrapSCIONPaths(dstAddr shila.NetworkAddress) ([]PathWrapper, error) {
	dstAddrIA := dstAddr.(*snet.UDPAddr).IA
	if paths, err := appnet.QueryPaths(dstAddrIA); err != nil {
		return nil, err
	} else if paths == nil {
		// Destination address is in the local IA
		return nil, nil
	} else {
		pathsWrapped := make([]PathWrapper, 0, len(paths))
		for _, path := range paths {
			rawMetrics := []int{int(path.MTU()), len(path.Interfaces())}
			pathsWrapped = append(pathsWrapped, PathWrapper{path: path, nUsed: 0, rawMetrics: rawMetrics })
			//log.Info.Printf("[%2d] %s\n", i, fmt.Sprintf("%s", path))
		}
		return pathsWrapped, nil
	}
}