package cmd

import "fmt"

func newUnexpectedResponse(code int) error {
	return fmt.Errorf("unexpected response: %v", code)
}
