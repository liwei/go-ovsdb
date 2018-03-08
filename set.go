package ovsdb

import (
	"encoding/json"
	"errors"
)

const (
	// setMagic identify a OVSDB set
	setMagic = "set"
)

var (
	errNotSet       = errors.New("Not an OVSDB set")
	errNotStringSet = errors.New("Not a StringSet")
)

// Set represents a OVSDB set
// https://tools.ietf.org/html/rfc7047#section-5.1
// <set>
// Either an <atom>, representing a set with exactly one element, or
// a 2-element JSON array that represents a database set value.  The
// first element of the array must be the string "set", and the
// second element must be an array of zero or more <atom>s giving the
// values in the set.  All of the <atom>s must have the same type.
type Set struct {
	Values []Value
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
	var ovsSet [2]interface{}
	if err := json.Unmarshal(value, &ovsSet); err != nil {
		return err
	}
	// the first element must be "SetMagic"
	magic, ok := ovsSet[0].(string)
	if !ok || magic != setMagic {
		return errNotSet
	}
	// the second element must be json array
	values, ok := ovsSet[1].([]interface{})
	if !ok {
		return errNotSet
	}
	for _, value := range values {
		s.Values = append(s.Values, Value(value))
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
	ovsSet = append(ovsSet, setMagic)
	ovsSet = append(ovsSet, s.Values)
	return json.Marshal(ovsSet)
}

// StringSet is a Set with element of string type
type StringSet struct {
	Values []string
}

// UnmarshalJSON decode json into an OVSDB set
func (s *StringSet) UnmarshalJSON(value []byte) error {
	// OVSDB set is either a atomic value
	if value[0] != '[' {
		var atomic string
		if err := json.Unmarshal(value, &atomic); err != nil {
			return err
		}
		s.Values = []string{atomic}
		return nil
	}

	// or a 2-element JSON array
	var ovsSet [2]interface{}
	if err := json.Unmarshal(value, &ovsSet); err != nil {
		return err
	}
	// the first element must be "SetMagic"
	magic, ok := ovsSet[0].(string)
	if !ok || magic != setMagic {
		return errNotSet
	}
	// the second element must be string array
	values, ok := ovsSet[1].([]interface{})
	if !ok {
		return errNotSet
	}

	s.Values = make([]string, len(values))
	for _, value := range values {
		strValue, ok := value.(string)
		if !ok {
			return errNotStringSet
		}
		s.Values = append(s.Values, strValue)
	}

	return nil
}

// MarshalJSON encode StringSet s into json format
func (s StringSet) MarshalJSON() ([]byte, error) {
	// 1-element array encoded to scalar value
	if len(s.Values) == 1 {
		return json.Marshal(s.Values[0])
	}

	var ovsSet []interface{}
	ovsSet = append(ovsSet, setMagic)
	ovsSet = append(ovsSet, s.Values)
	return json.Marshal(ovsSet)
}

// TODO: add other concrete Set for each scalar type
// XXX: should use some kind of code generation
