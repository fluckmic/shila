//
package shila

import "fmt"

// We dont need to terminate shila.
type TolerableError string
func (te TolerableError) Error() string {
	return string(te)
}

// We terminate shila, often an issue with the logic.
type CriticalError string
func (ce CriticalError) Error() string {
	return string(ce)
}

func PrependError(err error, msg string) error {
	switch err := err.(type) {
	case TolerableError:  			return TolerableError(fmt.Sprint(msg, err.Error()))
	case CriticalError:				return CriticalError(fmt.Sprint(msg, err.Error()))
	default:						return CriticalError(fmt.Sprint(msg, err.Error()))
	}
}