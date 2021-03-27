package server

import (
	"context"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/multiverse-vcs/go-git-ipfs/storage"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

// Loader loads repositories from IPFS.
type Loader struct {
	ctx context.Context
	ds  ipld.DAGService
}

// NewLoader returns a new IPFS loader.
func NewLoader(ctx context.Context, ds ipld.DAGService) *Loader {
	return &Loader{ctx, ds}
}

// Load loads a storer.Storer given a transport.Endpoint.
// Returns transport.ErrRepositoryNotFound if the repository does not
// exist.
func (l *Loader) Load(ep *transport.Endpoint) (storer.Storer, error) {
	parts := strings.Split(ep.Path, "/")

	id, err := cid.Decode(parts[1])
	if err != nil {
		return nil, err
	}

	fs, err := unixfs.Load(l.ctx, l.ds, id)
	if err != nil {
		return nil, err
	}

	return storage.NewStorage(fs), nil
}
