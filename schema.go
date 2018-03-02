package ovsdb

import (
	"encoding/json"
	"fmt"
	"io"
)

// DatabaseSchema represents the schema of a ovsdb database
type DatabaseSchema struct {
	// Name identifies the database as a whole
	Name ID `json:"name"`
	// Version reports the version of the database schema
	Version Version `json:"version"`
	// Checksum optionally reports an implementation-defined checksum for the database schema
	Checksum string `json:"cksum,omitempty"`
	// Tables is a JSON object whose names are table names and whose values are <table-schema>s.
	Tables map[ID]*TableSchema `json:"tables"`
}

// ColumnSet is an array of one or more strings,each of which names a column.
// Each Columnset is a set of columns whose values, taken together within any given row, must be
// unique within the table
type ColumnSet []string

// TableSchema represents the schema of a table in database
type TableSchema struct {
	// Columns is a JSON object whose names are column names and whose values are <column-schema>s.
	Columns map[ID]*ColumnSchema `json:"columns"`
	// MaxRows if specified, as a positive integer, it limits the maximum number of rows that may be present in the table
	MaxRows int `json:"maxRows,omitempty"`
	// IsRoot is used to determine whether rows in the table require strong references from other rows to avoid garbage collection
	IsRoot bool `json:"isRoot,omitempty"`
	// Indexes if specified, it must be an array of zero or more <ColumnSet>s
	Indexes []ColumnSet `json:"indexes,omitempty"`
}

// ColumnSchema represents the schema of a column in table
type ColumnSchema struct {
	// Type specifies the type of data stored in this column
	Type AtomicOrJSONColumnType `json:"type"`
	// Ephemeral if specified as true, then this column's values are not guaranteed to be durable
	Ephemeral bool `json:"ephemeral,omitempty"`
	// Mutable if specified as false, then this column's values may not be modified after they are initially set with the "insert" operation
	Mutable bool `json:"mutable,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler
// OVSDB has a wired default value of mutable, so need a custom json unmarshal function to set this default value
func (cs *ColumnSchema) UnmarshalJSON(value []byte) error {
	type aliasColumnSchema ColumnSchema
	alias := aliasColumnSchema{
		Mutable: true,
	}
	_ = json.Unmarshal(value, &alias)
	*cs = ColumnSchema(alias)
	return nil
}

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

// Dump writes the schema of the DatabaseSchema to io.Writer
func (dbSchema *DatabaseSchema) Dump(w io.Writer) {
	fmt.Fprintf(w, "%s (version: %q, checksum: %q)\n", dbSchema.Name, dbSchema.Version, dbSchema.Checksum)
	for table, tableSchema := range dbSchema.Tables {
		fmt.Fprintf(w, "\t %s (maxRows: %d, isRoot: %v)\n", table, tableSchema.MaxRows, tableSchema.IsRoot)
		for column, columnSchema := range tableSchema.Columns {
			fmt.Fprintf(w, "\t\t %s(ephemeral: %v, mutable: %v)\n", column, columnSchema.Ephemeral, columnSchema.Mutable)
			fmt.Fprintf(w, "\t\t %v\n", columnSchema)
		}
	}
}
