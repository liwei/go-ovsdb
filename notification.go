package ovsdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/cenkalti/rpc2"
)

// rpc2 doesn't support use method functions as message handlers,
// we need this to map a rpc2.Client to ovsdb.Client
// FIXME: fix rpc2 message handling or replace rpc2 with other jsonrpc client
var (
	clientsMap  map[*rpc2.Client]*Client
	clientsLock sync.RWMutex
)

// an empty NotificationHandlerFunc as default notification handler
var defaultNotificationHandler NotificationHandlerFuncs

// NotificationHandler is the interface for notification handlers to implement
type NotificationHandler interface {
	// Update notification is sent by the server to the client to report changes in tables that are being monitored
	Update(jsonValue Value, updates TableUpdates) error
	// Locked notification is provided to notify a client that it has been granted a lock that it had previously requested with the Lock method
	Locked(lock ID) error
	// Stolen notification is provided to notify a client, which had previously obtained a lock, that another client has stolen ownership of that lock
	Stolen(lock ID) error
}

// NotificationHandlerFuncs is a adapter which implements NotificationHandler interface
type NotificationHandlerFuncs struct {
	UpdateFunc func(jsonValue Value, updates TableUpdates) error
	LockedFunc func(lock ID) error
	StolenFunc func(lock ID) error
}

// TableUpdates is an object that maps from a table name to a TableUpdate
type TableUpdates map[ID]TableUpdate

// TableUpdate is an object that maps from the row's UUID to a RowUpdate object
type TableUpdate map[UUID]RowUpdate

// RowUpdate is an object with the following members:
// "old": <row>   present for "delete" and "modify" updates
// "new": <row>   present for "initial", "insert", and "modify" updates
type RowUpdate struct {
	Old Row `json:"old,omitempty"`
	New Row `json:"new,omitempty"`
}

// Update implements NotificationHandler interface
func (nh *NotificationHandlerFuncs) Update(jsonValue Value, updates TableUpdates) error {
	if nh.UpdateFunc == nil {
		return nil
	}
	return nh.UpdateFunc(jsonValue, updates)
}

// Locked implements NotificationHandler interface
func (nh *NotificationHandlerFuncs) Locked(lock ID) error {
	if nh.LockedFunc == nil {
		return nil
	}
	return nh.LockedFunc(lock)
}

// Stolen implements NotificationHandler interface
func (nh *NotificationHandlerFuncs) Stolen(lock ID) error {
	if nh.StolenFunc == nil {
		return nil
	}
	return nh.StolenFunc(lock)
}

// handler function for "update" notification
func updateHandler(client *rpc2.Client, params []interface{}, reply *[]interface{}) error {
	// "params": [<json-value>, <table-updates>]
	if len(params) != 2 {
		return errors.New("invalid update notification: wrong number of parameters")
	}

	var jsonValue = Value(params[0])
	var tableUpdates TableUpdates
	bytes, _ := json.Marshal(params[1])
	err := json.Unmarshal(bytes, &tableUpdates)
	if err != nil {
		return fmt.Errorf("failed to decode <table-updates>: %v", err)
	}

	clientsLock.RLock()
	ovsClient, ok := clientsMap[client]
	clientsLock.RUnlock()
	if ok {
		return ovsClient.handler.Update(jsonValue, tableUpdates)
	}
	return nil
}

// handler function for "locked" notification
func lockedHandler(client *rpc2.Client, params []interface{}, reply *[]interface{}) error {
	// "params": [<id>]
	// <id> is the lock name requested with a former lock method
	if len(params) != 1 {
		return errors.New("invalid locked notification: wrong number of parameters")
	}
	lock, ok := params[0].(string)
	if !ok {
		return errors.New("invalid locked notification: wrong lock name")
	}

	clientsLock.RLock()
	ovsClient, ok := clientsMap[client]
	clientsLock.RUnlock()
	if ok {
		return ovsClient.handler.Locked(ID(lock))
	}
	return nil
}

// handler function for "stolen" function
func stolenHandler(client *rpc2.Client, params []interface{}, reply *[]interface{}) error {
	// "params": [<id>]
	// <id> is the lock name which was stolen by another client
	if len(params) != 1 {
		return errors.New("invalid stolen notification: wrong number of parameters")
	}
	lock, ok := params[0].(string)
	if !ok {
		return errors.New("invalid stolen notification: wrong lock name")
	}

	clientsLock.RLock()
	ovsClient, ok := clientsMap[client]
	clientsLock.RUnlock()
	if ok {
		return ovsClient.handler.Stolen(ID(lock))
	}
	return nil
}
