package text

import (
	"fmt"
)

// ParseError contains details about a failed text parser match.
type ParseError struct {
	Parser string
	Input  string
	Line   int
}

func (e *ParseError) Error() string {
	location := ""
	if e.Line > 0 {
		location = fmt.Sprintf(" at line %d", e.Line)
	}

	return fmt.Sprintf("%s: unable to parse input%s: %q", e.Parser, location, e.Input)
}

func (e *ParseError) WithLineNumber(line int) *ParseError {
	e.Line = line
	return e
}

func newParseError(parser string, input string) *ParseError {
	return &ParseError{
		Parser: parser,
		Input:  input,
	}
}
