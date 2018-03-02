package ovsdb

import (
	"encoding/json"
	"errors"
)

const (
	// mapMagic identify a OVSDB map
	mapMagic = "map"
)

var (
	errNotMap = errors.New("Not an OVSDB map")
)

// Map represents an OVSDB map
// It's  2-element JSON array that represents a database map value.  The
// first element of the array must be the string "map", and the
// second element must be an array of zero or more <pair>s giving the
// values in the map.  All of the <pair>s must have the same key and
// value types.
// https://tools.ietf.org/html/rfc7047#section-5.1
type Map struct {
	Values []MapPair
}

// MapPair represents a pair within a OVSDB map
// <pair>
// A 2-element JSON array that represents a pair within a database
// map.  The first element is an <atom> that represents the key, and
// the second element is an <atom> that represents the value.
type MapPair [2]Atomic

// MarshalJSON implements json.Marshaler
func (m Map) MarshalJSON() ([]byte, error) {
	var ovsMap []interface{}
	ovsMap = append(ovsMap, mapMagic)
	ovsMap = append(ovsMap, m.Values)

	return json.Marshal(ovsMap)
}

// UnmarshalJSON implements json.Unmarshaler
func (m *Map) UnmarshalJSON(value []byte) error {
	var ovsMap [2]interface{}
	if err := json.Unmarshal(value, &ovsMap); err != nil {
		return err
	}
	magic, ok := ovsMap[0].(string)
	if !ok || magic != mapMagic {
		return errNotMap
	}
	// the second element must be json array
	values, ok := ovsMap[1].([]interface{})
	if !ok {
		return errNotMap
	}

	for _, value := range values {
		pair, ok := value.([]interface{})
		if !ok {
			return errNotMap
		}
		// MapPair must be a 2-element JSON array
		if len(pair) != 2 {
			return errNotMap
		}
		m.Values = append(m.Values, MapPair{pair[0], pair[1]})
	}
	return nil
}
