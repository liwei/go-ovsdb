package ovsdb

import (
	"encoding/json"
	"testing"
)

func TestSetMarshal(t *testing.T) {
	tests := []struct {
		set     Set
		jsonStr string
	}{
		{set: Set{Values: []Value{}}, jsonStr: `["set",[]]`},
		{set: Set{Values: []Value{"singleValue"}}, jsonStr: `"singleValue"`},
		{set: Set{Values: []Value{"strValue1", "strValue2"}}, jsonStr: `["set",["strValue1","strValue2"]]`},
		{set: Set{Values: []Value{1, 2, 3}}, jsonStr: `["set",[1,2,3]]`},
	}

	var bytes []byte
	var err error
	for _, test := range tests {
		bytes, err = json.Marshal(test.set)
		if err != nil {
			t.Errorf("Error during marshal: %v", err)
		}
		if string(bytes) != test.jsonStr {
			t.Errorf("json.Marshal(%+v) = %s, want %s", test.set, bytes, test.jsonStr)
		}
	}
}

func TestSetUnmarshal(t *testing.T) {
	tests := []struct {
		jsonStr string
		ok      bool
	}{
		{`["set",[]]`, true},
		{`"singleValue"`, true},
		{`["set",["strValue1","strValue2"]]`, true},
		{`["not", "2", "elements"]`, false},
		{`["set", "second element not a array"]`, false},
		{`["notset",["magic","is","not","set"]]`, false},
	}

	var set Set
	var err error
	for _, test := range tests {
		err = json.Unmarshal([]byte(test.jsonStr), &set)
		if test.ok && err != nil {
			t.Errorf("Error during unmarshal: %v", err)
		}
		if !test.ok && err == nil {
			t.Error("Expect error, got nil")
		}
	}
}
