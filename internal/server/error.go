package server

import "fmt"

type ServerHealthError struct {
	address string
}

func (s ServerHealthError) Error() string {
	return fmt.Sprintf("server %s is offline")
}
