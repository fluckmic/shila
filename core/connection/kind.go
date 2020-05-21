package connection

type kind uint8
const (
	_         = iota
	main kind = iota
	sub
	unknown
)

func(k kind) String() string {
	switch k {
	case main: 		return "main"
	case sub:  		return "sub"
	case unknown:   return "unknown"
	}
	return "unknown"
}
