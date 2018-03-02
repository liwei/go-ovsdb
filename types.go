package ovsdb

import "fmt"

// ID is a JSON string matching [a-zA-Z_][a-zA-Z0-9_]*. <id>s that begin
// with _ are reserved to the implementation and MUST NOT be used by
// the user.
type ID string

// Version is a JSON string that contains a version number that matches [0-9]+
// \.[0-9]+\.[0-9]+
type Version string

// Error is a struct to represents a ovsdb error
type Error struct {
	// Error is a short string that broadly indicates the class of the error
	Err string `json:"error"`
	// Details is a string that describes the error in more detail for the benefit of a human user or administrator
	Details string `json:"details,omitempty"`
}

// Error implements error interface
func (err *Error) Error() string {
	return fmt.Sprintf("%s(%s)", err.Err, err.Details)
}
