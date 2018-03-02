package ovsdb

import (
	"errors"
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
func (c *Client) Transact(db ID, ops ...Operation) error {
	if len(ops) < 1 {
		// 1 operations is required at least
		return errors.New("Not enough operation supplied, 1 at least")
	}

	var params []interface{}
	params = append(params, db)
	for _, op := range ops {
		params = append(params, op)
	}

	// FIXME: improve transact result processing
	var result []*Error
	c.rpc.Call("transact", params, &result)
	for _, r := range result {
		if len(r.Err) != 0 {
			return r
		}
	}

	return nil
}
