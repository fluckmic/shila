package netflow

import "fmt"

// Parsing issue.
type ParsingError string
func (e ParsingError) Error() string {
	return string(e)
}

type GeneralError string
func (e GeneralError) Error() string {
	return string(e)
}

func PrependError(err error, msg string) error {
	switch err := err.(type) {
	case ParsingError: 			return ParsingError(fmt.Sprint(msg, " - ", err.Error()))
	default:					return GeneralError(fmt.Sprint(msg, " - ", err.Error()))
	}
}