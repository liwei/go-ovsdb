package ovsdb

import (
	"encoding/json"
	"errors"
)

const (
	// Magic string to identify a OVSDB set
	SetMagic = "set"
)

var (
	ErrNotSet = errors.New("Not an OVSDB set")
)

///////////////////////////////////////////////////////////
// Types used in schema
//////////////////////////////////////////////////////////

// AtomicOrJSONColumnType is the type of a database column.  Either an <atomic-type> or a JSON
// object that describes the type of a database column
type AtomicOrJSONColumnType struct {
	IsAtomic bool
	Atomic   AtomicType
	JSON     JSONColumnType
}

// UnmarshalJSON implements json.Unmarshaler
func (atomjson *AtomicOrJSONColumnType) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		atomjson.IsAtomic = true
		return json.Unmarshal(value, &atomjson.Atomic)
	}
	atomjson.IsAtomic = false
	return json.Unmarshal(value, &atomjson.JSON)
}

// AtomicType is one of the strings "integer", "real", "boolean", "string", or "uuid", representing the specified scalar type.
type AtomicType string

// JSONColumnType is a JSON object that describes the type of a database column
type JSONColumnType struct {
	Key   AtomicOrJSONBaseType `json:"key"`
	Value AtomicOrJSONBaseType `json:"value,omitempty"`
	Min   int                  `json:"min,omitempty"`
	Max   IntOrString          `json:"max,omitempty"`
}

// IntOrString is a type that can hold an int or a string.  When used in
// JSON or YAML marshalling and unmarshalling, it produces or consumes the
// inner type.  This allows you to have, for example, a JSON field that can
// accept a name or number.
type IntOrString struct {
	IsInt bool
	Int   int
	Str   string
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (intstr *IntOrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		intstr.IsInt = false
		return json.Unmarshal(value, &intstr.Str)
	}
	intstr.IsInt = true
	return json.Unmarshal(value, &intstr.Int)
}

// AtomicOrJSONBaseType is the type of a key or value in a database column.  Either an
// <atomic-type> or a JSON object
type AtomicOrJSONBaseType struct {
	IsAtomic bool
	Atomic   AtomicType
	JSON     JSONBaseType
}

// UnmarshalJSON implements json.Unmarshaler
func (atomjson *AtomicOrJSONBaseType) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		atomjson.IsAtomic = true
		return json.Unmarshal(value, &atomjson.Atomic)
	}
	atomjson.IsAtomic = false
	return json.Unmarshal(value, &atomjson.JSON)
}

// JSONBaseType is a JSON object that describes the type of key or value
type JSONBaseType struct {
	Type       AtomicType `json:"type"`
	Enum       Set        `json:"enum,omitempty"`
	MinInteger int        `json:"minInteger,omitempty"`
	MaxInteger int        `json:"maxInteger,omitempty"`
	MinReal    float64    `json:"minReal,omitempty"`
	MaxReal    float64    `json:"maxReal,omitempty"`
	MinLength  int        `json:"minLength,omitempty"`
	MaxLength  int        `json:"maxLength,omitempty"`
	RefTable   ID         `json:"refTable,omitempty"`
	RefType    string     `json:"refType,omitempty"`
}

// Set represents a OVSDB set
// https://tools.ietf.org/html/rfc7047#section-5.1
// <set>
// Either an <atom>, representing a set with exactly one element, or
// a 2-element JSON array that represents a database set value.  The
// first element of the array must be the string "set", and the
// second element must be an array of zero or more <atom>s giving the
// values in the set.  All of the <atom>s must have the same type.
type Set struct {
	Values []interface{}
}

// UnmarshalJSON decode json into an OVSDB set
func (s *Set) UnmarshalJSON(value []byte) error {
	// OVSDB set is either a atomic value
	if value[0] != '[' {
		var atomic interface{}
		if err := json.Unmarshal(value, &atomic); err != nil {
			return err
		}
		s.Values = append(s.Values, atomic)
		return nil
	}

	// or a 2-element JSON array
	var ovsSet []interface{}
	if err := json.Unmarshal(value, &ovsSet); err != nil {
		return err
	}
	// must have 2 elements
	if len(ovsSet) != 2 {
		return ErrNotSet
	}
	// the first element must be "SetMagic"
	magic, ok := ovsSet[0].(string)
	if !ok || magic != SetMagic {
		return ErrNotSet
	}
	// the second element must be json array
	s.Values, ok = ovsSet[1].([]interface{})
	if !ok {
		return ErrNotSet
	}

	return nil
}

// MarshalJSON encode OVSDB set into json format
func (s Set) MarshalJSON() ([]byte, error) {
	// 1-element array encoded to scalar value
	if len(s.Values) == 1 {
		return json.Marshal(s.Values[0])
	}

	var ovsSet []interface{}
	ovsSet = append(ovsSet, SetMagic)
	ovsSet = append(ovsSet, s.Values)
	return json.Marshal(ovsSet)
}
