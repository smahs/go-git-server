package storage

import (
	"errors"
	"fmt"
	"strings"

	lerrors "github.com/syndtr/goleveldb/leveldb/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// SetReference implements the storer.ReferenceStorer interface
func (r *RepoStore) SetReference(ref *plumbing.Reference) error {
	var content string
	switch ref.Type() {
	case plumbing.SymbolicReference:
		content = fmt.Sprintf("ref: %s\n", ref.Target())
	case plumbing.HashReference:
		content = fmt.Sprintf(ref.Hash().String())
	}

	refName := ref.Name().String()
	return r.put([]byte(refName), []byte(content))
}

// CheckAndSetReference implements the storer.ReferenceStorer interface
func (r *RepoStore) CheckAndSetReference(new, old *plumbing.Reference) error {
	if err := r.delete([]byte("key")); err != nil {
		return err
	}
	return r.SetReference(new)
}

// Reference implements the storer.ReferenceStorer interface
func (r *RepoStore) Reference(name plumbing.ReferenceName) (*plumbing.Reference, error) {
	if name == plumbing.HEAD {
		return plumbing.NewSymbolicReference(
			plumbing.HEAD,
			plumbing.Master,
		), nil
	}

	var b, err = r.get([]byte(name.String()))
	if err != nil {
		if err == lerrors.ErrNotFound {
			return nil, plumbing.ErrReferenceNotFound
		}
		return nil, err
	}

	var line = strings.TrimSpace(string(b))
	return plumbing.NewReferenceFromStrings(name.String(), line), nil
}

// IterReferences implements the storer.ReferenceStorer interface
func (r *RepoStore) IterReferences() (storer.ReferenceIter, error) {
	var refs = []*plumbing.Reference{}
	var pairs, err = r.iter([]byte("refs"))
	if err != nil {
		return nil, err
	}
	for _, kv := range pairs {
		var key, val = string(kv.key), string(kv.value)
		refs = append(refs, plumbing.NewReferenceFromStrings(key, val))
	}
	return storer.NewReferenceSliceIter(refs), nil
}

// RemoveReference implements the storer.ReferenceStorer interface
func (r *RepoStore) RemoveReference(n plumbing.ReferenceName) error {
	return r.delete([]byte(n.String()))
}

// CountLooseRefs implements the storer.ReferenceStorer interface
func (r *RepoStore) CountLooseRefs() (int, error) {
	var pairs, err = r.iter([]byte("refs"))
	if err != nil {
		return 0, err
	}
	return len(pairs), nil
}

// PackRefs implements the storer.ReferenceStorer interface
func (r *RepoStore) PackRefs() error {
	return errors.New("Not implemented")
}
