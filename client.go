package ovsdb

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
)

// Client is a OVSDB client
type Client struct {
	rpc     *rpc2.Client
	schemas map[string]*DatabaseSchema
}

// Dial create a ovsdb.Client and connect to OVSDB server at address
func Dial(address string) (*Client, error) {
	var conn net.Conn
	var err error

	segs := strings.SplitN(address, ":", 2)
	switch segs[0] {
	case "tcp":
		conn, err = net.Dial("tcp", segs[1])
	case "unix":
		conn, err = net.Dial("unix", segs[1])
	default:
		return nil, fmt.Errorf("unknown protocol: %q", segs[0])
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	client := &Client{
		rpc:     rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(conn)),
		schemas: make(map[string]*DatabaseSchema),
	}

	// handle "echo" request from ovsdb-server, otherwise connection will be closed by server
	client.rpc.Handle("echo", func(client *rpc2.Client, args []interface{}, reply *[]interface{}) error {
		*reply = args
		return nil
	})
	// start rpc handling thread
	go client.rpc.Run()

	return client, nil
}

// ListDbs list databases in the connected OVSDB server
func (c *Client) ListDbs() ([]ID, error) {
	var dbs []ID
	if err := c.rpc.Call("list_dbs", nil, &dbs); err != nil {
		return nil, err
	}
	return dbs, nil
}

// GetSchema get the schema of a OVSDB database
func (c *Client) GetSchema(db ID) (*DatabaseSchema, error) {
	var dbSchema DatabaseSchema
	if err := c.rpc.Call("get_schema", db, &dbSchema); err != nil {
		return nil, err
	}
	return &dbSchema, nil
}

// Transact do operations as a transact on OVSDB
// https://tools.ietf.org/html/rfc7047#section-4.1.3
func (c *Client) Transact(db ID, ops ...Operation) (*TransactResult, error) {
	var result TransactResult
	// no operations supplied, return
	if len(ops) == 0 {
		return &result, nil
	}
	// construct rpc call parameters
	var params []interface{}
	params = append(params, db)
	for _, op := range ops {
		params = append(params, op)
	}

	err := c.rpc.Call("transact", params, &result)
	return &result, err
}

// TransactResult contains results for each operations in a transaction.
// See https://tools.ietf.org/html/rfc7047#section-4.1.3 for detailed explaination of the result array.
// For a failed operation, we decode the erorr message into ovsdb.Error, otherwise we keep the result
// as a json.RawMessage for user to decode it as proper operation result type.
type TransactResult struct {
	// Results contain operations' result
	Results []interface{}
	// At least one error in Results
	HasError bool
}

// UnmarshalJSON implements json.Unmarshaler interface
func (tr *TransactResult) UnmarshalJSON(value []byte) error {
	var raws []json.RawMessage
	// unmarshal into a RawMessage slice
	err := json.Unmarshal(value, &raws)
	if err != nil {
		return err
	}

	var temp map[string]interface{}
	for _, raw := range raws {
		err = json.Unmarshal(raw, &temp)
		if err != nil {
			return err
		}
		if temp == nil {
			// the operation was not attempted because a prior operation failed
			tr.Results = append(tr.Results, nil)
		} else if _, ok := temp["error"]; ok {
			// the operation completed with an error
			tr.HasError = true
			tr.Results = append(tr.Results, &Error{Err: temp["error"].(string), Details: temp["details"].(string)})
		} else {
			// the operation completed successfully
			tr.Results = append(tr.Results, raw)
		}
	}

	return nil
}
