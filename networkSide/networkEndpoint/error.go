package networkEndpoint

// An error related to the connection of a network endpoint.
type ConnectionError string
func (e ConnectionError) Error() string {
	return string(e)
}