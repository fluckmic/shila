package networkSide

import "shila/core/shila"

var _ shila.NetworkPath = (*Path)(nil)

type Path struct{}

func newPath(path string) shila.NetworkPath {
	// No path functionality w/ plain TCP.
	_ = path; return Path{}
}

func (p Path) String() string {
	// No path functionality w/ plain TCP.
	return ""
}
