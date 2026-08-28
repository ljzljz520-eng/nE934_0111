package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/flow026"
	"tastinginvite/internal/httpapi"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/store"
	"tastinginvite/internal/workflow"
)

func main() {
	path := flag.String("db", "invitations.db", "database path")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()
	database, err := store.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	clk := clock.System{}
	repo := repository.New(database, clk)
	engine := workflow.New(repo, clk)
	exporter := flow026.New(repo, clk)
	server := httpapi.New(repo, engine, exporter)
	log.Printf("private tasting invitation service listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.Mux); err != nil && !os.IsTimeout(err) {
		log.Fatal(err)
	}
}
