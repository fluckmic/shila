//
package shila

import "fmt"

// Most likely not our fault.
type ThirdPartyError string
func (tpe ThirdPartyError) Error() string {
	return string(tpe)
}

// Should not happen but we dont have to terminate shila.
// Could be our fault, but not necessarily.
type TolerableError string
func (te TolerableError) Error() string {
	return string(te)
}

// Terminate the shila, running it any further makes no sense.
// Most Issue with the implementation
type CriticalError string
func (ce CriticalError) Error() string {
	return string(ce)
}

// Parsing issue.
type ParsingError string
func (pe ParsingError) Error() string {
	return string(pe)
}

// Timeout issue in a network endpoint
type NetworkEndpointTimeout string
func (to NetworkEndpointTimeout) Error() string {
	return string(to)
}

func PrependError(err error, msg string) error {
	switch err := err.(type) {
	case ThirdPartyError: 			return ThirdPartyError(fmt.Sprint(msg, " - ", err.Error()))
	case TolerableError:  			return TolerableError(fmt.Sprint(msg, " - ", err.Error()))
	case ParsingError:				return ParsingError(fmt.Sprint(msg, " - ", err.Error()))
	case NetworkEndpointTimeout:	return NetworkEndpointTimeout(fmt.Sprint(msg, " - ", err.Error()))
	default:						return CriticalError(fmt.Sprint(msg, " - ", err.Error()))
	}
}