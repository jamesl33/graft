package rest

import "fmt"

// UnexpectedStatusCodeError is returned if a REST request resulted in a non-2xx status code.
type UnexpectedStatusCodeError struct {
	Status  int
	Method  string
	Address string
	Body    []byte
}

// Error implements the 'error' interface and returns a useful error message indicating what request failed.
func (e UnexpectedStatusCodeError) Error() string {
	msg := fmt.Sprintf("unexpected status code %d for %q request to %q", e.Status, e.Method, e.Address)
	if len(e.Body) != 0 {
		msg += fmt.Sprintf(", %s", e.Body)
	}

	return msg
}
