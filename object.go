package ipfs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/objfile"
	"github.com/go-git/go-git/v5/plumbing/storer"
	ufs "github.com/ipfs/go-unixfs"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

var ErrUnsupportedObjectType = errors.New("unsupported object type")

type ObjectStorage struct {
	fs *unixfs.Unixfs
}

func NewObjectStorage(fs *unixfs.Unixfs) ObjectStorage {
	return ObjectStorage{fs}
}

// NewEncodedObject returns a new plumbing.EncodedObject, the real type
// of the object can be a custom implementation or the default one,
// plumbing.MemoryObject.
func (o *ObjectStorage) NewEncodedObject() plumbing.EncodedObject {
	return &plumbing.MemoryObject{}
}

// SetEncodedObject saves an object into the storage, the object should
// be create with the NewEncodedObject, method, and file if the type is
// not supported.
func (o *ObjectStorage) SetEncodedObject(obj plumbing.EncodedObject) (plumbing.Hash, error) {
	if obj.Type() == plumbing.OFSDeltaObject || obj.Type() == plumbing.REFDeltaObject {
		return plumbing.ZeroHash, plumbing.ErrInvalidType
	}

	suffix := obj.Hash().String()[:2]
	prefix := obj.Hash().String()[2:]
	fpath := path.Join(unixfs.ObjectsPath, suffix, prefix)

	or, err := obj.Reader()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	defer or.Close()

	var b bytes.Buffer

	ow := objfile.NewWriter(&b)
	if err := ow.WriteHeader(obj.Type(), obj.Size()); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err := io.Copy(ow, or); err != nil {
		return plumbing.ZeroHash, err
	}

	if err := ow.Close(); err != nil {
		return plumbing.ZeroHash, err
	}

	if err := o.fs.Write(fpath, &b); err != nil {
		return plumbing.ZeroHash, err
	}

	return obj.Hash(), nil
}

// HasEncodedObject returns ErrObjNotFound if the object doesn't
// exist.  If the object does exist, it returns nil.
func (o *ObjectStorage) HasEncodedObject(h plumbing.Hash) error {
	suffix := h.String()[:2]
	prefix := h.String()[2:]
	fpath := path.Join(unixfs.ObjectsPath, suffix, prefix)

	_, err := o.fs.Find(fpath)
	if err == os.ErrNotExist {
		return plumbing.ErrObjectNotFound
	}

	return err
}

// EncodedObjectSize returns the plaintext size of the encoded object.
func (o *ObjectStorage) EncodedObjectSize(h plumbing.Hash) (int64, error) {
	suffix := h.String()[:2]
	prefix := h.String()[2:]
	fpath := path.Join(unixfs.ObjectsPath, suffix, prefix)

	node, err := o.fs.Find(fpath)
	if err == os.ErrNotExist {
		return 0, plumbing.ErrObjectNotFound
	}

	if err != nil {
		return 0, err
	}

	size, err := node.Size()
	return int64(size), err
}

// EncodedObject gets an object by hash with the given
// plumbing.ObjectType. Implementors should return
// (nil, plumbing.ErrObjectNotFound) if an object doesn't exist with
// both the given hash and object type.
//
// Valid plumbing.ObjectType values are CommitObject, BlobObject, TagObject,
// TreeObject and AnyObject. If plumbing.AnyObject is given, the object must
// be looked up regardless of its type.
func (o *ObjectStorage) EncodedObject(t plumbing.ObjectType, h plumbing.Hash) (plumbing.EncodedObject, error) {
	suffix := h.String()[:2]
	prefix := h.String()[2:]
	fpath := path.Join(unixfs.ObjectsPath, suffix, prefix)

	dr, err := o.fs.Read(fpath)
	if err == os.ErrNotExist {
		return nil, plumbing.ErrObjectNotFound
	}

	if err != nil {
		return nil, err
	}
	defer dr.Close()

	or, err := objfile.NewReader(dr)
	if err != nil {
		return nil, err
	}
	defer or.Close()

	typ, size, err := or.Header()
	if err != nil {
		return nil, err
	}

	obj := plumbing.MemoryObject{}
	obj.SetType(typ)

	read, err := io.Copy(&obj, or)
	if err != nil {
		return nil, err
	}

	if int64(read) != size {
		return nil, errors.New("invalid header size")
	}

	if obj.Type() != t && t != plumbing.AnyObject {
		return nil, plumbing.ErrObjectNotFound
	}

	return &obj, nil
}

// Objects returns a list of all object hashes.
func (o *ObjectStorage) Objects() ([]plumbing.Hash, error) {
	var hashes []plumbing.Hash

	walk := func(fpath string, node *ufs.FSNode) error {
		if node.IsDir() {
			return nil
		}

		parts := strings.Split(fpath, "/")
		if parts[1] == unixfs.InfoPath {
			return nil
		}

		hash := plumbing.NewHash(parts[1] + parts[2])
		hashes = append(hashes, hash)
		return nil
	}

	if err := o.fs.Walk(unixfs.ObjectsPath, walk); err != nil {
		return nil, err
	}

	return hashes, nil
}

// IterObjects returns a custom EncodedObjectStorer over all the object
// on the storage.
//
// Valid plumbing.ObjectType values are CommitObject, BlobObject, TagObject,
func (o *ObjectStorage) IterEncodedObjects(t plumbing.ObjectType) (storer.EncodedObjectIter, error) {
	objects, err := o.Objects()
	if err != nil {
		return nil, err
	}

	return &EncodedObjectIter{o, t, objects}, nil
}

// EncodedObjectIter looks up objects by hash.
type EncodedObjectIter struct {
	o      *ObjectStorage
	t      plumbing.ObjectType
	series []plumbing.Hash
}

func (iter *EncodedObjectIter) Next() (plumbing.EncodedObject, error) {
	if len(iter.series) == 0 {
		return nil, io.EOF
	}

	h := iter.series[0]
	iter.series = iter.series[1:]

	obj, err := iter.o.EncodedObject(iter.t, h)
	if err == plumbing.ErrObjectNotFound {
		return iter.Next()
	}

	return obj, err
}

func (iter *EncodedObjectIter) ForEach(cb func(plumbing.EncodedObject) error) error {
	return storer.ForEachIterator(iter, cb)
}

func (iter *EncodedObjectIter) Close() {
	iter.series = []plumbing.Hash{}
}
