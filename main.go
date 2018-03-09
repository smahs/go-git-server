package main

import (
	"appgit/storage"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var dbPath = "tmp/db"
var addr *url.URL
var server *http.Server
var store *storage.Store

func main() {
	var err error
	store, err = storage.NewStore(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := runServer(); err != nil {
		log.Fatal("Server initiation failed")
	}

	waitForInterrupt()
}

func runServer() error {
	var err error

	addr, err = url.Parse("http://localhost:9000")
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:    addr.String(),
		Handler: initMux(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	return nil
}

// Graceful Shurdown on SIGINT (Ctrl-C)
func waitForInterrupt() {
	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Wait indefinitely until SIGINT is received.
	<-c

	// Create a cancel context with a deadline to wait for.
	var wait time.Duration = time.Second * 30
	var ctx, cancel = context.WithTimeout(context.Background(), wait)
	defer cancel()

	server.Shutdown(ctx)

	log.Println("shutting down")
	os.Exit(0)
}
