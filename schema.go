package ovsdb

import (
	"encoding/json"
	"fmt"
	"io"
)

// ID is a JSON string matching [a-zA-Z_][a-zA-Z0-9_]*. <id>s that begin
// with _ are reserved to the implementation and MUST NOT be used by
// the user.
type ID string

// Version is a JSON string that contains a version number that matches [0-9]+
// \.[0-9]+\.[0-9]+
type Version string

// Error is a struct to represents a ovsdb error
type Error struct {
	// Error is a short string that broadly indicates the class of the error
	Error string `json:"error"`
	// Details is a string that describes the error in more detail for the benefit of a human user or administrator
	Details string `json:"details,omitempty"`
}

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
	Columns map[ID]ColumnSchema `json:"columns"`
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

// Dump writes the schema of the DatabaseSchema to io.Writer
func (dbSchema DatabaseSchema) Dump(w io.Writer) {
	fmt.Fprintf(w, "%s (version: %q, checksum: %q)\n", dbSchema.Name, dbSchema.Version, dbSchema.Checksum)
	for table, tableSchema := range dbSchema.Tables {
		fmt.Fprintf(w, "\t %s (maxRows: %d, isRoot: %v)\n", table, tableSchema.MaxRows, tableSchema.IsRoot)
		for column, columnSchema := range tableSchema.Columns {
			fmt.Fprintf(w, "\t\t %s(ephemeral: %v, mutable: %v)\n", column, columnSchema.Ephemeral, columnSchema.Mutable)
			fmt.Fprintf(w, "\t\t %v\n", columnSchema)
		}
	}
}
