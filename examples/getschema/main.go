package main

import (
	"flag"
	"log"
	"os"

	ovsdb "github.com/liwei/go-ovsdb"
)

var (
	address string
)

const (
	DefaultAddress = "unix:/var/run/openvswitch/db.sock"
)

func main() {
	flag.StringVar(&address, "address", DefaultAddress, "OVSDB server address")
	flag.Parse()

	ovsClient, err := ovsdb.Dial(address)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	dbs, err := ovsClient.ListDbs()
	if err != nil {
		log.Fatalf("failed to ListDbs: %v", err)
	}

	for _, db := range dbs {
		schema, err := ovsClient.GetSchema(db)
		if err != nil {
			log.Fatalf("failed to GetSchema: %v", err)
		}
		schema.Dump(os.Stdout)
	}
}
