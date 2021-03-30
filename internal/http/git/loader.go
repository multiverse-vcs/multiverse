package git

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/multiverse-vcs/go-git-ipfs/pkg/storage"
	"github.com/multiverse-vcs/go-git-ipfs/pkg/storage/unixfs"
)

// Loader loads repositories from IPFS.
type Loader struct {
	ctx context.Context
	ds  ipld.DAGService
	id  cid.Cid
	fs  *unixfs.Unixfs
}

// NewLoader returns a new IPFS loader.
func NewLoader(ctx context.Context, ds ipld.DAGService, id cid.Cid) *Loader {
	return &Loader{
		ctx: ctx,
		ds:  ds,
		id:  id,
	}
}

// Load loads a storer given a transport endpoint.
func (l *Loader) Load(ep *transport.Endpoint) (storer.Storer, error) {
	fs, err := unixfs.Load(l.ctx, l.ds, l.id)
	if err != nil {
		return nil, err
	}

	l.fs = fs
	return storage.NewStorage(fs), nil
}

// Node returns the final unixfs node.
func (l *Loader) Node() (ipld.Node, error) {
	if l.fs == nil {
		return nil, nil
	}

	return l.fs.Node()
}
