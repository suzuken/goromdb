package main

import (
	"flag"
	"fmt"

	"github.com/yowcow/go-romdb/protocol"
	"github.com/yowcow/go-romdb/protocol/memcachedprotocol"
	"github.com/yowcow/go-romdb/server"
	"github.com/yowcow/go-romdb/store"
	"github.com/yowcow/go-romdb/store/bdbstore"
	"github.com/yowcow/go-romdb/store/jsonstore"
	memcachedb_bdb "github.com/yowcow/go-romdb/store/memcachedb/bdbstore"
)

func main() {
	var addr string
	var protoBackend string
	var storeBackend string
	var file string

	flag.StringVar(&addr, "addr", ":11211", "Address to bind to")
	flag.StringVar(&protoBackend, "proto", "memcached", "Protocol: memcached")
	flag.StringVar(&storeBackend, "store", "bdb", "Store: json, bdb, memcachedb-bdb")
	flag.StringVar(&file, "file", "./data/sample-bdb.db", "Data file")
	flag.Parse()

	proto, err := createProtocol(protoBackend)

	if err != nil {
		panic(err)
	}

	store, err := createStore(storeBackend, file)

	if err != nil {
		panic(err)
	}

	s := server.New("tcp", addr, proto, store)
	s.Start()
}

func createProtocol(protoBackend string) (protocol.Protocol, error) {
	switch protoBackend {
	case "memcached":
		return memcachedprotocol.New()
	default:
		return nil, fmt.Errorf("don't know how to handle protoc '%s'", protoBackend)
	}
}

func createStore(storeBackend, file string) (store.Store, error) {
	switch storeBackend {
	case "bdb":
		return bdbstore.New(file)
	case "json":
		return jsonstore.New(file)
	case "memcachedb-bdb":
		return memcachedb_bdb.New(file)
	default:
		return nil, fmt.Errorf("don't know how to handle store '%s'", storeBackend)
	}
}
