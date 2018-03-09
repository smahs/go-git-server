package main

import (
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var head *plumbing.Reference

func prepareInMemoryRepo() {
	var err error

	var rConf = &config.RemoteConfig{
		Name: "origin",
		URLs: []string{addr.String()},
	}

	testRepo, err = git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		log.Fatal(err)
	}

	_, err = testRepo.CreateRemote(rConf)
	if err != nil {
		log.Fatal(err)
	}
}

func TestInfoRefsFailure(t *testing.T) {
	var uri, _ = url.Parse(testServer.URL)
	uri.Path = "/owner/repo.git/info/refs"

	resp, err := http.Get(uri.String())
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404; received: %s", resp.Status)
	}

	uri.RawQuery = "service=git-service-pack"
	resp, err = http.Get(uri.String())
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400; received: %s", resp.Status)
	}

	uri.RawQuery = "service=git-upload-pack"
	resp, err = http.Post(uri.String(), "", nil)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405; received: %s", resp.Status)
	}
}

func TestInfoRefsEmpty(t *testing.T) {
	var uri, _ = url.Parse(testServer.URL)
	uri.Path = "/owner/repo.git/info/refs"
	uri.RawQuery = "service=git-upload-pack"

	resp, err := http.Get(uri.String())
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200; received: %s", resp.Status)
	}

	defer resp.Body.Close()

	var refs = packp.NewAdvRefs()
	refs.Decode(resp.Body)

	if len(refs.References) != 0 {
		t.Errorf("Expected 0 refs; found %d",
			len(refs.References))
	}
}

func TestReceivePack(t *testing.T) {
	prepareInMemoryRepo()
	wTree, err := testRepo.Worktree()
	file, err := wTree.Filesystem.Create("README.md")
	if err != nil {
		log.Fatal("Bad test! File creation failed")
	}
	file.Write([]byte("Hello World!"))

	if _, err = wTree.Add("README.md"); err != nil {
		log.Fatal("Could not add file to worktree")
	}

	commit, err := wTree.Commit("test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Fatal("Creating a new local commit failed")
	}

	if err = testRepo.Push(&git.PushOptions{}); err != nil {
		t.Errorf("Push operation failed")
	}

	head, err = testRepo.Head()
	if err != nil {
		t.Errorf("Undefined head: %s", err.Error())
	}
	if commit != head.Hash() {
		t.Errorf("Pull hash mismatch")
	}
}

func TestUploadPack(t *testing.T) {
	prepareInMemoryRepo()
	wTree, err := testRepo.Worktree()
	options := &git.PullOptions{RemoteName: "origin"}
	if err = wTree.Pull(options); err != nil {
		t.Errorf("Pull operation failed")
	}

	newHead, err := testRepo.Head()
	if err != nil {
		log.Fatal("Could not read HEAD after pull")
	}
	if newHead.Hash() != head.Hash() {
		t.Errorf("Push hash mismatch")
	}
}
