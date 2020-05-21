package model

type EntityStateIdentifier uint8

const (
	_                                   = iota
	Uninitialized EntityStateIdentifier = iota
	Initialized
	Running
	TornDown
)

func (es EntityStateIdentifier) String() string {
	switch es {
	case Uninitialized: 	return "Uninitialized"
	case Initialized: 		return "Initialized"
	case Running:			return "Running"
	case TornDown:			return "TornDown"
	default:				return "Unknown"
	}
}

type EntityState struct {
	EntityStateIdentifier
}

func(es *EntityState) Set(s EntityStateIdentifier) {
	es.EntityStateIdentifier = s
}

func (es *EntityState) Get() EntityStateIdentifier {
	return es.EntityStateIdentifier
}