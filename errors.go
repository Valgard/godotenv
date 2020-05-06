package godotenv

import (
	"fmt"
)

type ParseError struct {
	Line int
	Position int
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Line %d: %s", e.Line, e.Message)
}

type PathError struct {
	Path string
	Err error
}

func (e PathError) Error() string {
	return fmt.Sprintf("unable to read the %q environment file", e.Path)
}

func (e PathError) Unwrap() error {
	return e.Err
}
