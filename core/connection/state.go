package connection

type stateIdentifier uint8
const (
	_                           = iota
	established stateIdentifier = iota
	serverReady
	clientReady
	clientEstablished
	closed
	raw
)

func(s stateIdentifier) String() string {
	switch s {
	case established: 			return "Established"
	case serverReady: 			return "ServerReady"
	case clientReady: 			return "ClientReady"
	case clientEstablished:		return "ClientEstablished"
	case closed:				return "Closed"
	case raw:					return "Raw"
	}
	return "Unknown"
}

type state struct {
	previous stateIdentifier
	current  stateIdentifier
}

func (s *state) set(newState stateIdentifier) {
	s.previous = s.current
	s.current = newState
}

func newState() state {
	return state{previous: raw, current: raw}
}