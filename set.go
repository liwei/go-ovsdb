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
	errNotSet = errors.New("Not an OVSDB set")
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
		return errNotSet
	}
	// the first element must be "SetMagic"
	magic, ok := ovsSet[0].(string)
	if !ok || magic != setMagic {
		return errNotSet
	}
	// the second element must be json array
	s.Values, ok = ovsSet[1].([]interface{})
	if !ok {
		return errNotSet
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
