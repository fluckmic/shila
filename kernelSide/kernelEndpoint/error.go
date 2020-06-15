package kernelEndpoint

// An error related to the connection of a kernel endpoint.
type ConnectionError string
func (e ConnectionError) Error() string {
	return string(e)
}