//
package network

import "shila/core/shila"

// Generator functionalities are thought to be used outside of the
// backbone protocol specific implementations (suffix "Specific")
var _ shila.NetworkPathGenerator = (*PathGenerator)(nil)
var _ shila.NetworkPath 		 = (*Path)(nil)

type PathGenerator struct {}

func (g PathGenerator) New(path string) (shila.NetworkPath, error) {
	return newPath(path)
}

func newPath(p string) (shila.NetworkPath, error) {
	// No Path functionality w/ plain TCP.
	_ = p
	return Path{}, nil
}

func (g PathGenerator) NewEmpty() shila.NetworkPath {
	return Path{}
}

type Path struct{}

func (p Path) String() string {
	// No Path functionality w/ plain TCP.
	return ""
}
