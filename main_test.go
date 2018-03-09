package main

import (
	"appgit/storage"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	git "gopkg.in/src-d/go-git.v4"
)

var testServer *httptest.Server
var testRepo *git.Repository
var testDBDir = "db"
var testRepoPath = "/owner/repo.git"

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	shutdown()
	os.Exit(exitCode)
}

func setup() {
	log.Println("Starting the test server...")

	var err error
	if _, err = os.Stat(testDBDir); os.IsNotExist(err) {
		if err = os.Mkdir(testDBDir, 0777); err != nil {
			log.Fatal(err)
		}
	}

	store, err = storage.NewStore(testDBDir)
	if err != nil {
		log.Fatal(err)
	}

	var mux = initMux()
	testServer = httptest.NewServer(mux)

	addr, err = url.Parse(testServer.URL)
	if err != nil {
		log.Fatal(err)
	}
	addr.Path = testRepoPath
}

func shutdown() {
	log.Println("Stoping the test server...")
	testServer.Close()
	os.RemoveAll(testDBDir)
}
