package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/badger/v3"

	abciserver "github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/stevenferrer/kvstore-tm"
)

var (
	socketAddr string
	dbPath     string
)

func init() {
	flag.StringVar(&socketAddr, "socket-addr", "unix://example.sock", "Unix domain socket address")
	flag.StringVar(&dbPath, "db-path", "/tmp/badger", "DB file path")
}

func main() {
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger db: %v", err)
		os.Exit(1)
	}
	defer db.Close()
	app := kvstore.NewApplication(db)

	flag.Parse()

	logger, err := log.NewDefaultLogger(log.LogFormatPlain, log.LogLevelInfo, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure logger: %v", err)
		os.Exit(1)
	}

	server := abciserver.NewSocketServer(socketAddr, app)
	server.SetLogger(logger)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	defer server.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
