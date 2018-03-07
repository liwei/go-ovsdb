package ovsdb

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Operation represents a operation on OVSDB
// see: https://tools.ietf.org/html/rfc7047#section-5.2
type Operation interface {
	Op() OperationType
}

// OperationType is the type of operation
type OperationType string

// Operation types
const (
	OpInsert  OperationType = "insert"
	OpSelect  OperationType = "select"
	OpUpdate  OperationType = "update"
	OpMutate  OperationType = "mutate"
	OpDelete  OperationType = "delete"
	OpWait    OperationType = "wait"
	OpCommit  OperationType = "commit"
	OpAbort   OperationType = "abort"
	OpComment OperationType = "comment"
	OpAssert  OperationType = "assert"
)

/////////////////////////////////////////////////////////////////////
// insert operation
// https://tools.ietf.org/html/rfc7047#section-5.2.1
/////////////////////////////////////////////////////////////////////

// InsertOperation insert Row into Table
type InsertOperation struct {
	Table    ID
	Row      Row
	UUIDName ID
}

// MarshalJSON implements json.Marshaler interface
func (insert InsertOperation) MarshalJSON() ([]byte, error) {
	// validate required fields
	switch {
	case len(insert.Table) == 0:
		return nil, errors.New("Table field is required")
	case insert.Row == nil:
		return nil, errors.New("Row field is required")
	}

	var temp = struct {
		Op       OperationType `json:"op"`
		Table    ID            `json:"table"`
		Row      Row           `json:"row"`
		UUIDName ID            `json:"uuid-name,omitempty"`
	}{
		Op:       OpInsert,
		Table:    insert.Table,
		Row:      insert.Row,
		UUIDName: insert.UUIDName,
	}

	return json.Marshal(temp)
}

// Op implements Operation interface
func (insert *InsertOperation) Op() OperationType {
	return OpInsert
}

/////////////////////////////////////////////////////////////////////
// mutate operation
// https://tools.ietf.org/html/rfc7047#section-5.2.4
/////////////////////////////////////////////////////////////////////

// MutateOperation mutates rows that match all the conditions specified in Where in Table
type MutateOperation struct {
	Table     ID
	Where     []Condition
	Mutations []Mutation
}

// MarshalJSON implements json.Marshaler interface
func (mutate MutateOperation) MarshalJSON() ([]byte, error) {
	// validate required fields
	switch {
	case len(mutate.Table) == 0:
		return nil, errors.New("Table field is required")
	case len(mutate.Where) == 0:
		return nil, errors.New("Where field is required")
	case len(mutate.Mutations) == 0:
		return nil, errors.New("Mutations field is required")
	}
	// validate contions
	for _, cond := range mutate.Where {
		if !cond.Valid() {
			return nil, fmt.Errorf("Invalid condition: %v", cond)
		}
	}
	// validate mutations
	for _, mutation := range mutate.Mutations {
		if !mutation.Valid() {
			return nil, fmt.Errorf("Invalid mutation: %v", mutation)
		}
	}

	var temp = struct {
		Op        OperationType `json:"op"`
		Table     ID            `json:"table"`
		Where     []Condition   `json:"where"`
		Mutations []Mutation    `json:"mutations"`
	}{
		Op:        OpMutate,
		Table:     mutate.Table,
		Where:     mutate.Where,
		Mutations: mutate.Mutations,
	}

	return json.Marshal(temp)
}

// Op implements Operation interface
func (mutate *MutateOperation) Op() OperationType {
	return OpMutate
}

// Condition is a 3-element JSON array of the form [<column>, <function>, <value>]
// that represents a test on a column value.
type Condition struct {
	Column   ID
	Function Function
	Value    Value
}

// MarshalJSON implements json.Marshaler interface
func (c Condition) MarshalJSON() ([]byte, error) {
	var temp []interface{}
	temp = append(temp, c.Column)
	temp = append(temp, c.Function)
	temp = append(temp, c.Value)

	return json.Marshal(temp)
}

// Valid returns true if condition is valid, otherwise false
func (c Condition) Valid() bool {
	// TODO: pass in a ColumnSchema and do validation based on it
	switch c.Function {
	case FuncLt, FuncLe, FuncEq, FuncNe, FuncGt, FuncGe, FuncInc, FuncExc:
		return true
	}
	return false
}

// Function is the condition operator
// It is one of "<", "<=", "==", "!=", ">=", ">", "includes", or "excludes"
// and supported mutators depend on the type of column
type Function string

// Functions supported in Condition
const (
	FuncLt  Function = "<"
	FuncLe  Function = "<="
	FuncEq  Function = "=="
	FuncNe  Function = "!="
	FuncGt  Function = ">"
	FuncGe  Function = ">="
	FuncInc Function = "includes"
	FuncExc Function = "excludes"
)

// Mutation is a 3-element JSON array of the form [<column>, <mutator>, <value>]
// that represents a change to a column value.
type Mutation struct {
	Column  ID
	Mutator Mutator
	Value   Value
}

// MarshalJSON implements json.Marshaler interface
func (m Mutation) MarshalJSON() ([]byte, error) {
	var temp []interface{}
	temp = append(temp, m.Column)
	temp = append(temp, m.Mutator)
	temp = append(temp, m.Value)

	return json.Marshal(temp)
}

// Valid returns true if mutation is valid, otherwise false
func (m Mutation) Valid() bool {
	// TODO: pass in a ColumnSchema and do validation based on it
	switch m.Mutator {
	case MutatorPluEq, MutatorMinEq, MutatorMulEq, MutatorDivEq, MutatorModEq, MutatorInsert, MutatorDelete:
		return true
	}
	return false
}

// Mutator define the mutation operation on column
// It is one of "+=", "-=", "*=", "/=", "%=", "insert", or "delete"
// and supported mutators depend on the type of column
type Mutator string

// Mutators supported in Mutation
const (
	MutatorPluEq  Mutator = "+="
	MutatorMinEq  Mutator = "-="
	MutatorMulEq  Mutator = "*="
	MutatorDivEq  Mutator = "/="
	MutatorModEq  Mutator = "%="
	MutatorInsert Mutator = "insert"
	MutatorDelete Mutator = "delete"
)
