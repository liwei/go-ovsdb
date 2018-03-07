package ovsdb

import (
	"encoding/json"
	"testing"
)

func TestInsertOperation(t *testing.T) {
	insertOp := &InsertOperation{}
	if op := insertOp.Op(); op != OpInsert {
		t.Errorf("Op() returned %q, want \"insert\"", op)
	}
	marshalTests := []struct {
		op         InsertOperation
		shouldFail bool
		json       string
	}{
		{InsertOperation{}, true, ``},
		{InsertOperation{Row: map[ID]Value{"TestColumn": "TestValue"}}, true, ``},
		{InsertOperation{Table: "TestTable"}, true, ``},
		{InsertOperation{Table: "TestTable", Row: map[ID]Value{"TestColumn": "TestValue"}}, false, `{"op":"insert","table":"TestTable","row":{"TestColumn":"TestValue"}}`},
		{InsertOperation{Table: "TestTable", Row: map[ID]Value{"TestColumn": "TestValue"}, UUIDName: "TestUUIDName"}, false, `{"op":"insert","table":"TestTable","row":{"TestColumn":"TestValue"},"uuid-name":"TestUUIDName"}`},
	}
	for _, test := range marshalTests {
		bytes, err := json.Marshal(test.op)
		if test.shouldFail {
			if err == nil {
				t.Error("expect json marshal failed, but got nil")
			}
			continue
		}
		if err != nil {
			t.Error("json marshal failed")
		}
		if string(bytes) != test.json {
			t.Errorf("json marshal got %q, want %q", bytes, test.json)
		}
	}
}
