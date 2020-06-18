package router

import "shila/core/shila"

type paths struct {

}

func newPaths(dstAddr shila.NetworkAddress) paths {
	return paths{}
}

func (p *paths) get(key shila.IPFlowKey) shila.NetworkPath {
	return nil
}

func (p *paths) free(key shila.IPFlowKey) {}


