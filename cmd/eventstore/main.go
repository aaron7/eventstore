package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aaron7/eventstore/pkg/db"
	"github.com/aaron7/eventstore/pkg/store"
)

func main() {
	var (
		listen = flag.String("listen", ":8000", "listen address")
		dbPath = flag.String("db", "badger://.db", "db path e.g. badger://.db or memory://")
		debug  = flag.Bool("debug", false, "Enable debug endpoints")
	)
	flag.Parse()

	db, err := db.New(*dbPath)
	if err != nil {
		return
	}
	defer db.Close()

	s, err := store.New(db)
	if err != nil {
		return
	}

	api := &store.API{
		Store: s,
		Debug: *debug,
	}

	http.Handle("/", api)
	http.Handle("/metrics", promhttp.Handler())

	fmt.Printf("Listening on %s\n", *listen)
	http.ListenAndServe(*listen, nil)
}
