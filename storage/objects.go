package storage

import (
	"bytes"
	"errors"
	"io/ioutil"
	"regexp"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

var regex = regexp.MustCompile("^/objects/(.*?)/type$")

// NewEncodedObject implements the storer.EncodedObjectStorer interface
func (r *RepoStore) NewEncodedObject() plumbing.EncodedObject {
	return &plumbing.MemoryObject{}
}

// SetEncodedObject implements the storer.EncodedObjectStorer interface
func (r *RepoStore) SetEncodedObject(obj plumbing.EncodedObject) (plumbing.Hash, error) {
	var key = append([]byte("objects/" + obj.Hash().String()))

	reader, err := obj.Reader()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	err = r.put(append(key, []byte("/type")...), []byte(obj.Type().String()))
	if err != nil {
		return plumbing.ZeroHash, err
	}

	err = r.put(key, content)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return obj.Hash(), err
}

// EncodedObject implements the storer.EncodedObjectStorer interface
func (r *RepoStore) EncodedObject(t plumbing.ObjectType,
	h plumbing.Hash) (plumbing.EncodedObject, error) {
	var typ plumbing.ObjectType
	var content []byte

	var pairs, err = r.iter([]byte("objects"))
	if err != nil {
		return nil, err
	}

	for _, p := range pairs {
		var key = p.key
		if !bytes.Contains(key, []byte(h.String())) {
			continue
		}
		if bytes.HasSuffix(key, []byte("type")) {
			var value = string(p.value)
			typ, err = plumbing.ParseObjectType(value)
			if err != nil {
				return nil, err
			}
			continue
		}
		content = p.value
	}

	return objectFromBytes(content, typ)
}

func objectFromBytes(c []byte, t plumbing.ObjectType) (plumbing.EncodedObject, error) {
	var o = &plumbing.MemoryObject{}
	o.SetType(t)
	o.SetSize(int64(len(c)))

	_, err := o.Write(c)
	if err != nil {
		return nil, err
	}

	return o, nil
}

// IterEncodedObjects implements the storer.EncodedObjectStorer interface
func (r *RepoStore) IterEncodedObjects(t plumbing.ObjectType) (storer.EncodedObjectIter, error) {
	var types = make(map[plumbing.ObjectType][]string)
	var values = make(map[string][]byte)
	var pairs, err = r.iter([]byte("objects"))
	if err != nil {
		return nil, err
	}

	for _, p := range pairs {
		var key = p.key
		if bytes.HasSuffix(key, []byte("type")) {
			var value = string(p.value)
			var obType, err = plumbing.ParseObjectType(value)
			if err != nil {
				return nil, err
			}
			types[obType] = append(
				types[obType],
				regex.FindStringSubmatch(string(key))[1],
			)
			continue
		}
		values[string(key)] = p.value
	}

	var objects []plumbing.EncodedObject
	for _, hash := range types[t] {
		obj, err := objectFromBytes(values[hash], t)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}

	return storer.NewEncodedObjectSliceIter(objects), nil
}

// HasEncodedObject implements the storer.EncodedObjectStorer interface
func (r *RepoStore) HasEncodedObject(h plumbing.Hash) error {
	c, err := r.get([]byte("objects/" + h.String()))
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("Not exists")
	}
	return nil
}
