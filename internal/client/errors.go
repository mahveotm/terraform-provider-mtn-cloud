package client

import "fmt"

type APIError struct {
	StatusCode int
	Message    string
	Body       map[string]any
}

func (e *APIError) Error() string {
	if e.StatusCode == 0 {
		return e.Message
	}
	return fmt.Sprintf("mtn cloud api error %d: %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == 404
}
