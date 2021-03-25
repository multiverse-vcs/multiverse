package ipfs

import (
	"io"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage"
	ufs "github.com/ipfs/go-unixfs"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

type ReferenceStorage struct {
	fs *unixfs.Unixfs
}

func NewReferenceStorage(fs *unixfs.Unixfs) ReferenceStorage {
	return ReferenceStorage{fs}
}

func (r *ReferenceStorage) SetReference(ref *plumbing.Reference) error {
	parts := ref.Strings()
	name, target := parts[0], parts[1]
	return r.fs.Write(name, strings.NewReader(target))
}

// CheckAndSetReference sets the reference `new`, but if `old` is
// not `nil`, it first checks that the current stored value for
// `old.Name()` matches the given reference value in `old`.  If
// not, it returns an error and doesn't update `new`.
func (r *ReferenceStorage) CheckAndSetReference(ref, old *plumbing.Reference) error {
	if old == nil {
		return r.SetReference(ref)
	}

	tmp, err := r.Reference(old.Name())
	if err != nil {
		return storage.ErrReferenceHasChanged
	}

	if tmp.Hash() != old.Hash() {
		return storage.ErrReferenceHasChanged
	}

	return r.SetReference(ref)
}

func (r *ReferenceStorage) Reference(n plumbing.ReferenceName) (*plumbing.Reference, error) {
	fr, err := r.fs.Read(n.String())
	if err == os.ErrNotExist {
		return nil, plumbing.ErrReferenceNotFound
	}

	if err != nil {
		return nil, err
	}
	defer fr.Close()

	target, err := io.ReadAll(fr)
	if err != nil {
		return nil, err
	}

	return plumbing.NewReferenceFromStrings(n.String(), string(target)), nil
}

func (r *ReferenceStorage) IterReferences() (storer.ReferenceIter, error) {
	refs, err := r.References()
	if err != nil {
		return nil, err
	}

	return storer.NewReferenceSliceIter(refs), nil
}

func (r *ReferenceStorage) References() ([]*plumbing.Reference, error) {
	var refs []*plumbing.Reference

	head, err := r.Reference("HEAD")
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return nil, err
	}

	if err == nil {
		refs = append(refs, head)
	}

	walk := func(fpath string, node *ufs.FSNode) error {
		if node.IsDir() {
			return nil
		}

		ref, err := r.Reference(plumbing.ReferenceName(fpath))
		if err != nil {
			return err
		}

		refs = append(refs, ref)
		return nil
	}

	if err := r.fs.Walk(unixfs.RefsPath, walk); err != nil {
		return nil, err
	}

	return refs, nil
}

func (r *ReferenceStorage) CountLooseRefs() (int, error) {
	refs, err := r.References()
	if err != nil {
		return 0, err
	}

	return len(refs), nil
}

func (r *ReferenceStorage) PackRefs() error {
	return nil
}

func (r *ReferenceStorage) RemoveReference(n plumbing.ReferenceName) error {
	err := r.fs.Remove(n.String())
	if err == os.ErrNotExist {
		return nil
	}

	return err
}
