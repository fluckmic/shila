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
	case established: 			return "established"
	case serverReady: 			return "serverReady"
	case clientReady: 			return "clientReady"
	case clientEstablished:		return "clientEstablished"
	case closed:				return "closed"
	case raw:					return "raw"
	}
	return "unknown"
}

type state struct {
	previous stateIdentifier
	current  stateIdentifier
}

func (s *state) Set(newState stateIdentifier) {
	s.previous = s.current
	s.current = newState
}
