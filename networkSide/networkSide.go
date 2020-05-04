package networkSide

import (
	"shila/networkSide/networkEndpoint"
)

type Manager struct {
	Endpoints map[string]*networkEndpoint.Generator
}

type Error string

func (e Error) Error() string {
	return string(e)
}
