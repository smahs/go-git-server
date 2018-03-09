package storage

import (
	"bytes"
	"log"

	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Store holds the underlying LevelDB connection
// Its safe to use on concurrent goroutines
type Store struct {
	db *leveldb.DB
}

// NewStore returns a new Store instance
func NewStore(path string) (*Store, error) {
	o := &opt.Options{
		Filter: filter.NewBloomFilter(16),
	}
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		log.Fatal(err)
	}

	return &Store{db}, nil
}

// NewRepoStore instantiates a RepoStore for a repo name
// A repo name is conventionally similar to most Git systems
// For example, `/owner/repo.git`
func (s *Store) NewRepoStore(name string) *RepoStore {
	return &RepoStore{[]byte(name), s}
}

// RepoStore is a wrapper for Store, but restricted to
// operations for a single repository path
// Results can be at best unpredictrable, if receive-pack
// and upload-pack operations are performed concurrently
type RepoStore struct {
	name  []byte
	store *Store
}

// formatKey prepares the path for given key appended to reponame
// For example, `refs` => `/owner/repo.git/refs`
func (r *RepoStore) formatKey(key []byte) []byte {
	var sep = []byte("/")
	return append(append(r.name, sep...), key...)
}

// stripKey removes the repo name from the key
// For example, `/owner/repo.git/refs` => `refs`
func (r *RepoStore) stripKey(key []byte) []byte {
	var sep = []byte("/")
	var pre = append(r.name, sep...)
	return bytes.TrimPrefix(key, pre)
}

// Load implements transport/server.Loader interface
func (r *RepoStore) Load(ep *transport.Endpoint) (storer.Storer, error) {
	return r, nil
}

// put saves the key value in the db
func (r *RepoStore) put(key, value []byte) error {
	return r.store.db.Put(r.formatKey(key), value, nil)
}

// get gets the value for the given key, if it exists
// Returns leveldb/errors.ErrNotFound if key does not exist
func (r *RepoStore) get(key []byte) ([]byte, error) {
	return r.store.db.Get(r.formatKey(key), nil)
}

// delete deletes the key from the db
func (r *RepoStore) delete(key []byte) error {
	return r.store.db.Delete(r.formatKey(key), nil)
}

// kvPair type holds the kev/value pair retrieved from leveldb
type kvPair struct {
	key, value []byte
}

// iter iterates through keys/values for given prefix
// For example, to get all objects, r.iter(]byte("objects"))
func (r *RepoStore) iter(pre []byte) ([]kvPair, error) {
	pre = r.formatKey(pre)
	var pairs []kvPair
	var iter = r.store.db.NewIterator(util.BytesPrefix(pre), nil)
	defer iter.Release()
	for iter.Next() {
		key := iter.Key()
		var keyb = make([]byte, len(key))
		copy(keyb, key)

		value := iter.Value()
		var valb = make([]byte, len(value))
		copy(valb, value)

		pairs = append(pairs, kvPair{r.stripKey(keyb), valb})
	}
	return pairs, iter.Error()
}
