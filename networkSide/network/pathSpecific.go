package network

import "shila/core/shila"

// Generator functionalities are thought to be used outside of the
// backbone protocol specific implementations (suffix "Specific")
var _ shila.NetworkPathGenerator = (*PathGenerator)(nil)
var _ shila.NetworkPath 		 = (*path)(nil)

type PathGenerator struct {}

func (g PathGenerator) New(path string) shila.NetworkPath {
	return newPath(path)
}

func newPath(p string) shila.NetworkPath {
	// No path functionality w/ plain TCP.
	_ = p; return path{}
}

func (g PathGenerator) NewEmpty() shila.NetworkPath {
	return path{}
}

type path struct{}

func (p path) String() string {
	// No path functionality w/ plain TCP.
	return ""
}
