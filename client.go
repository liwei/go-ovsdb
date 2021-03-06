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
	handler NotificationHandler
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
		handler: &defaultNotificationHandler,
	}

	// insert this client to clientsMap
	clientsLock.Lock()
	if clientsMap == nil {
		clientsMap = make(map[*rpc2.Client]*Client)
	}
	clientsMap[client.rpc] = client
	clientsLock.Unlock()

	// handle "echo" request from ovsdb-server, otherwise connection will be closed by server
	client.rpc.Handle("echo", echoHandler)
	// register notification handlers
	client.rpc.Handle("update", updateHandler)
	client.rpc.Handle("locked", lockedHandler)
	client.rpc.Handle("stolen", stolenHandler)

	// start rpc handling thread
	go client.rpc.Run()

	return client, nil
}

func echoHandler(client *rpc2.Client, args []interface{}, reply *[]interface{}) error {
	*reply = args
	return nil
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
	// Errors keeps operation errors in a separate slice for convenience
	Errors ResultErrors
}

// ResultErrors is a slice of Error that can be treat as a single error
type ResultErrors []*Error

// Error implements error interface
func (re ResultErrors) Error() string {
	errMsgs := []string{}
	for _, err := range re {
		errMsgs = append(errMsgs, err.Err)
	}
	return strings.Join(errMsgs, ", ")
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
			opError := &Error{
				Err:     temp["error"].(string),
				Details: temp["details"].(string),
			}
			tr.Errors = append(tr.Errors, opError)
			tr.Results = append(tr.Results, opError)
		} else {
			// the operation completed successfully
			tr.Results = append(tr.Results, raw)
		}
	}

	return nil
}

// SetNotificationHandler set handler as the notification handler
// FIXME: not thread-safe
func (c *Client) SetNotificationHandler(handler NotificationHandler) {
	c.handler = handler
}

// Monitor enables a client to replicate tables or subsets
// of tables within an OVSDB database by requesting notifications of
// changes to those tables and by receiving the complete initial state
// of a table or a subset of a table
func (c *Client) Monitor(db ID, jsonValue Value, requests MonitorRequests) (TableUpdates, error) {
	var updates TableUpdates
	params := []interface{}{db, jsonValue, requests}
	if err := c.rpc.Call("monitor", params, &updates); err != nil {
		return nil, err
	}
	return updates, nil
}

// MonitorRequests maps the name of the table to be monitored to an array of MonitorRequest
type MonitorRequests map[ID]MonitorRequest

// MonitorRequest selects the contents to monitor in a table
type MonitorRequest struct {
	// Columns, if present, define the columns within the table to be monitored,
	// if omitted, all columns in the table, except for "_uuid", are monitored.
	Columns []ID           `json:"columns,omitempty"`
	Select  *MonitorSelect `json:"select,omitempty"`
}

// MonitorSelect specify how the columns or table are to be monitored
type MonitorSelect map[SelectType]bool

// SelectType is the type of MonitorSelect, valid values are: "initial", "insert", "delete", "modify"
type SelectType string

// Supported SelectTypes
const (
	SelectInitial = "initial"
	SelectInsert  = "insert"
	SelectDelete  = "delete"
	SelectModify  = "modify"
)

// MonitorCancel cancels a previously issued monitor request
func (c *Client) MonitorCancel(jsonValue Value) error {
	return c.rpc.Call("monitor_cancel", []interface{}{jsonValue}, nil)
}

// Lock acquire a lock named lockID from OVSDB server
func (c *Client) Lock(lockID ID) (bool, error) {
	var result LockResult
	if err := c.rpc.Call("lock", []interface{}{lockID}, &result); err != nil {
		return false, err
	}
	return result.Locked, nil
}

// LockResult is the result of Lock method
type LockResult struct {
	Locked bool `json:"locked"`
}

// Steal acquire a lock named lockID from OVSDB server.
// If there is an existing owner, it loses ownership.
func (c *Client) Steal(lockID ID) error {
	return c.rpc.Call("steal", []interface{}{lockID}, nil)
}

// Unlock release a lock named lockID
func (c *Client) Unlock(lockID ID) error {
	return c.rpc.Call("unlock", []interface{}{lockID}, nil)
}
