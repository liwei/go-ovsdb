package main

import (
	"flag"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"time"

	goovsdb "github.com/liwei/go-ovsdb"
)

const (
	rpcServerAddress = "192.168.122.153:6641"
	rpcDialTimeout   = 5 * time.Second
)

var (
	ovsdb string
)

func main() {
	flag.StringVar(&ovsdb, "ovsdb", "192.168.122.153:6641", "Database path to connect to")
	flag.Parse()

	conn, err := net.DialTimeout("tcp", rpcServerAddress, rpcDialTimeout)
	if err != nil {
		log.Fatalf("failed to dial to rpc server: %v", err)
	}
	defer conn.Close()
	// create jsonrpc client
	client := jsonrpc.NewClient(conn)

	dbs := []string{}
	err = client.Call("list_dbs", nil, &dbs)
	if err != nil {
		log.Fatalf("failed to call list_dbs: %v", err)
	}
	if len(dbs) == 0 {
		log.Fatal("no database found in ovsdbserver")
	}

	var dbSchema goovsdb.DatabaseSchema
	err = client.Call("get_schema", dbs[0], &dbSchema)
	if err != nil {
		log.Fatalf("failed to get database schema: %v", err)
	}

	dbSchema.Dump(os.Stdout)
}
