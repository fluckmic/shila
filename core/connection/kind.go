package connection

type kind uint8
const (
	_             = iota
	mainflow kind = iota
	subflow
	unknown
)

func(k kind) String() string {
	switch k {
	case mainflow: 		return "MainFlow"
	case subflow:  		return "SubFlow"
	case unknown:   	return "Unknown"
	}
	return "Unknown"
}
