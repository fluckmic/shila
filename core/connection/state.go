package connection

type stateIdentifier uint8
const (
	_                           = iota
	Established stateIdentifier = iota
	ServerReady
	ClientReady
	ClientEstablished
	Closed
	Raw
)

func(s stateIdentifier) String() string {
	switch s {
	case Established: 			return "Established"
	case ServerReady: 			return "ServerReady"
	case ClientReady: 			return "ClientReady"
	case ClientEstablished:		return "ClientEstablished"
	case Closed:				return "Closed"
	case Raw:					return "Raw"
	}
	return "Unknown"
}

type state struct {
	previous stateIdentifier
	current  stateIdentifier
}

func (s *state) Set(newState stateIdentifier) {
	s.previous = s.current
	s.current = newState
}

func (s *state) Current() stateIdentifier {
	return s.current
}

func (s *state) Previous() stateIdentifier {
	return s.previous
}

type kind uint8
const (
	_             = iota
	MainFlow kind = iota
	SubFlow
	Unknown
)

func(c kind) String() string {
	switch c {
	case MainFlow: return "MainFlow"
	case SubFlow:  return "SubFlow"
	case Unknown:  return "Unknown"
	}
	return "Unknown"
}