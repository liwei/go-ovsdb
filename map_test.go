package ovsdb

import (
	"encoding/json"
	"testing"
)

func TestMapMarshal(t *testing.T) {
	tests := []struct {
		m       Map
		jsonStr string
	}{
		{m: Map{Values: []MapPair{}}, jsonStr: `["map",[]]`},
		{m: Map{Values: []MapPair{MapPair{"key", "value"}}}, jsonStr: `["map",[["key","value"]]]`},
		{m: Map{Values: []MapPair{MapPair{"key1", "value1"}, MapPair{"key2", "value2"}}}, jsonStr: `["map",[["key1","value1"],["key2","value2"]]]`},
		{m: Map{Values: []MapPair{MapPair{1, "value"}}}, jsonStr: `["map",[[1,"value"]]]`},
		{m: Map{Values: []MapPair{MapPair{"key", 1}}}, jsonStr: `["map",[["key",1]]]`},
		{m: Map{Values: []MapPair{MapPair{1, 2}}}, jsonStr: `["map",[[1,2]]]`},
	}

	var bytes []byte
	var err error
	for _, test := range tests {
		bytes, err = json.Marshal(test.m)
		if err != nil {
			t.Errorf("Error during marshal: %v", err)
		}
		if string(bytes) != test.jsonStr {
			t.Errorf("json.Marshal(%+v) = %s, want %s", test.m, bytes, test.jsonStr)
		}
	}

}

func TestMapUnmarshal(t *testing.T) {
	tests := []struct {
		jsonStr string
		ok      bool
	}{
		{`["map",[]]`, true},
		{`["map",[["key","value"]]]`, true},
		{`["map",[["key","value"],["key1","value1"]]]`, true},
		{`["map",[["intval",1]]]`, true},
		{`["map",[[1,"intkey"]]]`, true},
		{`["not", "2", "elements"]`, false},
		{`["map", "second element not a array"]`, false},
		{`["notmap",[["magic","is"],["not","map"]]]`, false},
		{`["map",["mappair not array"]]`, false},
		{`["map",[["not",2,"elements"]]]`, false},
	}

	var m Map
	var err error
	for _, test := range tests {
		err = json.Unmarshal([]byte(test.jsonStr), &m)
		if test.ok && err != nil {
			t.Errorf("Error during unmarshal: %v", err)
		}
		if !test.ok && err == nil {
			t.Error("Expect error, got nil")
		}
	}
}
