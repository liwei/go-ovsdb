package ovsdb

import (
	"encoding/json"
	"testing"
)

func TestInsertOperation(t *testing.T) {
	insertOp := &InsertOperation{}
	if op := insertOp.Op(); op != OpInsert {
		t.Errorf("Op() returned %q, want %q", op, OpInsert)
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

func TestSelectOperation(t *testing.T) {
	s := &SelectOperation{}
	if op := s.Op(); op != OpSelect {
		t.Errorf("Op() returned %q, want %q", op, OpSelect)
	}
	marshalTests := []struct {
		op         SelectOperation
		shouldFail bool
		json       string
	}{
		// missing required fields
		{SelectOperation{}, true, ``},
		{SelectOperation{Table: "TestTable"}, true, ``},
		// without columns
		{
			op: SelectOperation{
				Table: "TestTable",
				Where: []Condition{Condition{"TestColumn", "==", "TestValue"}},
			},
			shouldFail: false,
			json:       `{"op":"select","table":"TestTable","where":[["TestColumn","==","TestValue"]]}`,
		},
		// with columns
		{
			op: SelectOperation{
				Table:   "TestTable",
				Where:   []Condition{Condition{"TestColumn", "==", "TestValue"}},
				Columns: []ID{"TestColumn"},
			},
			shouldFail: false,
			json:       `{"op":"select","table":"TestTable","where":[["TestColumn","==","TestValue"]],"columns":["TestColumn"]}`,
		},
		// invalid condition
		{
			op: SelectOperation{
				Table: "TestTable",
				Where: []Condition{Condition{"TestColumn", "invalid function", "TestValue"}},
			},
			shouldFail: true,
			json:       ``,
		},
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

func TestUpdateOperation(t *testing.T) {
	u := &UpdateOperation{}
	if op := u.Op(); op != OpUpdate {
		t.Errorf("Op() returned %q, want %q", op, OpMutate)
	}
	marshalTests := []struct {
		op         UpdateOperation
		shouldFail bool
		json       string
	}{
		// empty
		{UpdateOperation{}, true, ``},
		// missing Table
		{UpdateOperation{Table: "TestTable"}, true, ``},
		{
			op: UpdateOperation{
				Where: []Condition{Condition{"TestColumn", "==", "TestValue"}},
				Row:   map[ID]Value{"TestColumn": "NewValue"},
			},
			shouldFail: true,
			json:       ``,
		},
		// Missing Where
		{UpdateOperation{Table: "TestTable"}, true, ``},
		{
			op: UpdateOperation{
				Table: "TestTable",
				Row:   map[ID]Value{"TestColumn": "NewValue"},
			},
			shouldFail: true,
			json:       ``,
		},
		// Missing Row
		{UpdateOperation{Table: "TestTable"}, true, ``},
		{
			op: UpdateOperation{
				Table: "TestTable",
				Where: []Condition{Condition{"TestColumn", "==", "TestValue"}},
			},
			shouldFail: true,
			json:       ``,
		},
		// valid case
		{
			op: UpdateOperation{
				Table: "TestTable",
				Where: []Condition{Condition{"TestColumn", "==", "TestValue"}},
				Row:   map[ID]Value{"TestColumn": "NewValue"},
			},
			shouldFail: false,
			json:       `{"op":"update","table":"TestTable","where":[["TestColumn","==","TestValue"]],"row":{"TestColumn":"NewValue"}}`,
		},
		// invalid condition
		{
			op: UpdateOperation{
				Table: "TestTable",
				Where: []Condition{Condition{"TestColumn", "invalid function", "TestValue"}},
				Row:   map[ID]Value{"TestColumn": "NewValue"},
			},
			shouldFail: true,
			json:       ``,
		},
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

func TestMutateOperation(t *testing.T) {
	mutateOp := &MutateOperation{}
	if op := mutateOp.Op(); op != OpMutate {
		t.Errorf("Op() returned %q, want %q", op, OpMutate)
	}
	marshalTests := []struct {
		op         MutateOperation
		shouldFail bool
		json       string
	}{
		// missing required fields
		{MutateOperation{}, true, ``},
		{MutateOperation{Table: "TestTable"}, true, ``},
		{
			op: MutateOperation{
				Table:     "TestTable",
				Where:     []Condition{Condition{"TestColumn", "==", "TestValue"}},
				Mutations: []Mutation{},
			},
			shouldFail: true,
			json:       ``,
		},
		// valid case
		{
			op: MutateOperation{
				Table:     "TestTable",
				Where:     []Condition{Condition{"TestColumn", "==", "TestValue"}},
				Mutations: []Mutation{Mutation{"TestColumn", "+=", 1}},
			},
			shouldFail: false,
			json:       `{"op":"mutate","table":"TestTable","where":[["TestColumn","==","TestValue"]],"mutations":[["TestColumn","+=",1]]}`,
		},
		// invalid condition
		{
			op: MutateOperation{
				Table:     "TestTable",
				Where:     []Condition{Condition{"TestColumn", "invalid function", "TestValue"}},
				Mutations: []Mutation{Mutation{"TestColumn", "+=", 1}},
			},
			shouldFail: true,
			json:       ``,
		},
		// invalid mutation
		{
			op: MutateOperation{
				Table:     "TestTable",
				Where:     []Condition{Condition{"TestColumn", "==", "TestValue"}},
				Mutations: []Mutation{Mutation{"TestColumn", "invalid mutator", 1}},
			},
			shouldFail: true,
			json:       ``,
		},
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
