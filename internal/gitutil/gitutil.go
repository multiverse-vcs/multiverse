package gitutil

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/multiverse-vcs/go-git-ipfs/pkg/storage"
	"github.com/multiverse-vcs/go-git-ipfs/pkg/storage/unixfs"
)

var readme = regexp.MustCompile(`(?i)^read\s*me(\..*)?$`)

// Init initializes a new repository and returns its unixfs node.
func Init(ctx context.Context, ds ipld.DAGService) (ipld.Node, error) {
	fs, err := unixfs.New(ctx, ds)
	if err != nil {
		return nil, err
	}

	if _, err = git.Init(storage.NewStorage(fs), nil); err != nil {
		return nil, err
	}

	return fs.Node()
}

// Open returns the git repository with the given CID.
func Open(ctx context.Context, ds ipld.DAGService, id string) (*git.Repository, error) {
	c, err := cid.Decode(id)
	if err != nil {
		return nil, err
	}

	fs, err := unixfs.Load(ctx, ds, c)
	if err != nil {
		return nil, err
	}

	return git.Open(storage.NewStorage(fs), nil)
}

// Branches returns a list of all repository branches.
func Branches(repo *git.Repository) ([]*plumbing.Reference, error) {
	iter, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	var branches []*plumbing.Reference
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref)
		return nil
	})

	return branches, err
}

// References returns a map of references and commits.
func References(repo *git.Repository) (map[plumbing.ReferenceName]*object.Commit, error) {
	iter, err := repo.References()
	if err != nil {
		return nil, err
	}

	refs := make(map[plumbing.ReferenceName]*object.Commit)
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		refs[ref.Name()] = commit
		return nil
	})

	return refs, err
}

// Logs returns a list of commits from the repo head.
func Logs(repo *git.Repository, ref *plumbing.Reference) ([]*object.Commit, error) {
	opts := git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	}

	iter, err := repo.Log(&opts)
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	err = iter.ForEach(func(commit *object.Commit) error {
		commits = append(commits, commit)
		return nil
	})

	return commits, err
}


// RefPath splits a path into the ref and path parts.
func RefPath(repo *git.Repository, path string) (*plumbing.Reference, string, error) {
	iter, err := repo.References()
	if err != nil {
		return nil, "", err
	}

	var ref *plumbing.Reference
	err = iter.ForEach(func(r *plumbing.Reference) error {
		if strings.HasPrefix(path, r.Name().String()) {
			ref = r
		}

		return nil
	})

	if ref == nil {
		return nil, "", errors.New("invalid ref path")
	}

	path = strings.TrimPrefix(path, ref.Name().String())
	return ref, path, err
}

// Find returns a tree or blob from the given repo at the given ref and path.
func Find(repo *git.Repository, ref *plumbing.Reference, path string) (object.Object, error) {
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	if path == "" {
		return tree, nil
	}

	entry, err := tree.FindEntry(path)
	if err != nil {
		return nil, err
	}

	switch {
	case entry.Mode.IsFile():
		return repo.BlobObject(entry.Hash)
	default:
		return repo.TreeObject(entry.Hash)
	}
}

// Readme returns the readme blob object if one exists.
func Readme(repo *git.Repository, ref *plumbing.Reference) (*object.Blob, error) {
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	for _, e := range tree.Entries {
		if e.Mode.IsFile() && readme.MatchString(e.Name) {
			return repo.BlobObject(e.Hash)
		}
	}

	return nil, nil
}
