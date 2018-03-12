package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/smahs/go-git-server/storage"
)

var dbPath = "tmp/db"
var addr *url.URL
var server *http.Server
var store *storage.Store
var wg sync.WaitGroup

func main() {
	var err error
	store, err = storage.NewStore(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	if err := runServer(); err != nil {
		log.Fatal("Server initiation failed")
	}

	wg.Add(1)
	go waitForInterrupt()

	wg.Wait()
}

func runServer() error {
	var err error

	addr, err = url.Parse("http://localhost:9000")
	if err != nil {
		return err
	}

	server = &http.Server{
		Addr:        addr.Host,
		Handler:     initMux(),
		IdleTimeout: time.Duration(10 * time.Second),
	}

	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	return nil
}

// Graceful Shutdown on SIGINT (Ctrl-C)
func waitForInterrupt() {
	defer wg.Done()

	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Wait indefinitely until SIGINT is received.
	<-c

	// Create a cancel context with a deadline to wait for.
	var wait time.Duration = time.Second * 30
	var ctx, cancel = context.WithTimeout(context.Background(), wait)
	defer cancel()

	server.Shutdown(ctx)
}
