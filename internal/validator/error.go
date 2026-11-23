package validator

import "fmt"

// Error carries field-level validation failures across package boundaries.
type Error struct {
	Fields map[string]string
}

func (e *Error) Error() string {
	return fmt.Sprintf("validation failed on %d field(s)", len(e.Fields))
}
