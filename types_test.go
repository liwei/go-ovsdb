package ovsdb

import (
	"encoding/json"
	"testing"
)

func TestUUIDMarshal(t *testing.T) {
	tests := []struct {
		uuid    UUID
		jsonStr string
	}{
		{uuid: "550e8400-e29b-41d4-a716-446655440000", jsonStr: `["uuid","550e8400-e29b-41d4-a716-446655440000"]`},
	}

	var bytes []byte
	var err error
	for _, test := range tests {
		bytes, err = json.Marshal(test.uuid)
		if err != nil {
			t.Errorf("Error during marshal: %v", err)
		}
		if string(bytes) != test.jsonStr {
			t.Errorf("json.Marshal(%+v) = %s, want %s", test.uuid, bytes, test.jsonStr)
		}
	}
}

func TestUUIDUnmarshal(t *testing.T) {
	tests := []struct {
		jsonStr string
		ok      bool
	}{
		{``, false},
		{`[]`, false},
		{`["not uuid"]`, false},
		{`["uuid","invalid length"]`, false},
		{`["uuid",1]`, false},
		{`["uuid","012345678901234567890123456789012345"]`, true},
	}

	var uuid UUID
	var err error
	for _, test := range tests {
		err = json.Unmarshal([]byte(test.jsonStr), &uuid)
		if test.ok && err != nil {
			t.Errorf("Error during unmarshal: %v", err)
		}
		if !test.ok && err == nil {
			t.Error("Expect error, got nil")
		}
	}

}

func TestNamedUUIDMarshal(t *testing.T) {
	tests := []struct {
		uuid    NamedUUID
		jsonStr string
	}{
		{uuid: "uuid", jsonStr: `["named-uuid","uuid"]`},
	}

	var bytes []byte
	var err error
	for _, test := range tests {
		bytes, err = json.Marshal(test.uuid)
		if err != nil {
			t.Errorf("Error during marshal: %v", err)
		}
		if string(bytes) != test.jsonStr {
			t.Errorf("json.Marshal(%+v) = %s, want %s", test.uuid, bytes, test.jsonStr)
		}
	}
}

func TestNamedUUIDUnmarshal(t *testing.T) {
	tests := []struct {
		jsonStr string
		ok      bool
	}{
		{``, false},
		{`[]`, false},
		{`["not named-uuid"]`, false},
		{`["invalid magic","uuid"]`, false},
		{`["named-uuid",1]`, false},
		{`["named-uuid","uuid"]`, true},
	}

	var uuid NamedUUID
	var err error
	for _, test := range tests {
		err = json.Unmarshal([]byte(test.jsonStr), &uuid)
		if test.ok && err != nil {
			t.Errorf("Error during unmarshal: %v", err)
		}
		if !test.ok && err == nil {
			t.Error("Expect error, got nil")
		}
	}

}
