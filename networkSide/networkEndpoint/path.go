package networkEndpoint

import "shila/core/model"

var _ model.NetworkPath = (*Path)(nil)

type Path struct{}

func newPath(path string) model.NetworkPath {
	// No path functionality w/ plain TCP.
	_ = path; return Path{}
}

func (p Path) String() string {
	// No path functionality w/ plain TCP.
	return ""
}
