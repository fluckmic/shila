package shila

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
	}
	return "Unknown"
}

type EntityState struct {
	EntityStateIdentifier
}

func NewEntityState() EntityState {
	return EntityState{EntityStateIdentifier: Uninitialized}
}

func(es *EntityState) Set(s EntityStateIdentifier) {
	es.EntityStateIdentifier = s
}

func (es *EntityState) Not(s EntityStateIdentifier) bool {
	return es.EntityStateIdentifier != s
}

func (es *EntityState) String() string {
	return es.EntityStateIdentifier.String()
}