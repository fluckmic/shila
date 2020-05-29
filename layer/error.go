//
package layer

// Most likely not our fault.
type ParsingError string
func (pe ParsingError) Error() string {
	return string(pe)
}
