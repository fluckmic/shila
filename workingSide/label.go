package workingSide

type Label uint8

const (
	_             = iota
	Ingress Label = iota
	Egress
)

func (l Label) String() string {
	switch l {
	case Ingress: 	return "Ingress"
	case Egress: 	return "Egress"
	}
	return "Unknown"
}