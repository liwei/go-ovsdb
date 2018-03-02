package ovsdb

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Magics to identify different OVSDB types
const (
	uuidMagic      = "uuid"
	namedUUIDMagic = "named-uuid"
)

var (
	errNotUUID      = errors.New("Not an OVSDB UUID")
	errNotNamedUUID = errors.New("Not an OVSDB NamedUUID")
)

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

// The following implements the simple types in RFC 7047
// see: https://tools.ietf.org/html/rfc7047#section-5.1
// Complex types (e.g. set and map) are in their separate files.

// Row represents a row in a OVSDB table
// <row>
// A JSON object that describes a table row or a subset of a table
// row.  Each member is the name of a table column paired with the
// <value> of that column.
type Row map[ID]Value

// Value is the value of a column
// <value>
// A JSON value that represents the value of a column in a table row,
// one of <atom>, <set>, or <map>.
// FIXME: define more concrete type instead of interface{}
type Value interface{}

// Atomic is a scalar value for a column
// <atom>
// A JSON value that represents a scalar value for a column, one of
// <string>, <number>, <boolean>, <uuid>, or <named-uuid>.
// FIXME: define more concrete type instead of interface{}
type Atomic interface{}

// UUID is a 2-element JSON array that represents a UUID
// The first element of the array must be the string "uuid", and the second element
// must be a 36-character string giving the UUID in the format described by RFC 4122
type UUID string

const uuidLen = 36

// MarshalJSON implements json.Marshaler interface
func (uuid UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{uuidMagic, string(uuid)})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (uuid *UUID) UnmarshalJSON(value []byte) error {
	var ovsUUID []string
	err := json.Unmarshal(value, &ovsUUID)
	if err != nil {
		return err
	}

	if len(ovsUUID) != 2 || ovsUUID[0] != uuidMagic || len(ovsUUID[1]) != uuidLen {
		return errNotUUID
	}

	*uuid = UUID(ovsUUID[1])
	return nil
}

// NamedUUID is a 2-element JSON array that represents the UUID of a row inserted
// in an "insert" operation within the same transaction
// The first element of the array must be the string "named-uuid", and the
// second element should be the <id> specified as the "uuid-name" for
// an "insert" operation within the same transaction.
type NamedUUID string

// MarshalJSON implements json.Marshaler interface
func (nu NamedUUID) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{namedUUIDMagic, string(nu)})
}

// UnmarshalJSON implements json.Unmarshaler interface
func (nu *NamedUUID) UnmarshalJSON(value []byte) error {
	var ovsNamedUUID []string
	err := json.Unmarshal(value, &ovsNamedUUID)
	if err != nil {
		return err
	}

	if len(ovsNamedUUID) != 2 || ovsNamedUUID[0] != namedUUIDMagic {
		return errNotUUID
	}

	*nu = NamedUUID(ovsNamedUUID[1])
	return nil
}
