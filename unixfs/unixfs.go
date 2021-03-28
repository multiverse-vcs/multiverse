package unixfs

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	cid "github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	ufs "github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer"
	ufsio "github.com/ipfs/go-unixfs/io"
)

const (
	ObjectsPath = "objects"
	RefsPath    = "refs"
	InfoPath    = "info"
	PackPath    = "pack"
)

// Unixfs is used to read and write to a unixfs directory.
type Unixfs struct {
	ctx context.Context
	ds  ipld.DAGService
	dir ufsio.Directory
}

// New returns a new unixfs directory initialized with empty directories.
func New(ctx context.Context, ds ipld.DAGService) (*Unixfs, error) {
	dir := ufsio.NewDirectory(ds)

	if err := ds.Add(ctx, ufs.EmptyDirNode()); err != nil {
		return nil, err
	}

	if err := dir.AddChild(ctx, ObjectsPath, ufs.EmptyDirNode()); err != nil {
		return nil, err
	}

	if err := dir.AddChild(ctx, RefsPath, ufs.EmptyDirNode()); err != nil {
		return nil, err
	}

	return &Unixfs{ctx, ds, dir}, nil
}

// Load returns an existing unixfs using the directory with the given id.
func Load(ctx context.Context, ds ipld.DAGService, id cid.Cid) (*Unixfs, error) {
	node, err := ds.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	dir, err := ufsio.NewDirectoryFromNode(ds, node)
	if err != nil {
		return nil, err
	}

	if err := ds.Add(ctx, ufs.EmptyDirNode()); err != nil {
		return nil, err
	}

	return &Unixfs{ctx, ds, dir}, nil
}

// Save writes the final unixfs to the dag service.
func (fs *Unixfs) Save() (cid.Cid, error) {
	node, err := fs.dir.GetNode()
	if err != nil {
		return cid.Cid{}, err
	}

	if err := fs.ds.Add(fs.ctx, node); err != nil {
		return cid.Cid{}, err
	}

	return node.Cid(), nil
}

// Module returns a new unixfs for a submodule
func (fs *Unixfs) Module() (*Unixfs, error) {
	return New(fs.ctx, fs.ds)
}

// Read returns a reader for the file at the given path.
func (fs *Unixfs) Read(fpath string) (io.ReadCloser, error) {
	node, err := fs.Find(fpath)
	if err != nil {
		return nil, err
	}

	r, err := ufsio.NewDagReader(fs.ctx, node, fs.ds)
	if err == ufsio.ErrIsDir {
		return nil, os.ErrNotExist
	}

	return r, nil
}

// Write writes the contents of the given reader to the path.
func (fs *Unixfs) Write(fpath string, r io.Reader) error {
	file, err := importer.BuildDagFromReader(fs.ds, chunker.DefaultSplitter(r))
	if err != nil {
		return err
	}

	return fs.writePath(fs.dir, strings.Split(fpath, "/"), file)
}

// Find returns the node at the given path.
func (fs *Unixfs) Find(fpath string) (ipld.Node, error) {
	return fs.findPath(fs.dir, strings.Split(fpath, "/"))
}

// Remove removes the node at the given path.
func (fs *Unixfs) Remove(fpath string) error {
	return fs.removePath(fs.dir, strings.Split(fpath, "/"))
}

// WalkFun is called for each directory entry.
type WalkFun func(string, *ufs.FSNode) error

// Walk walks the directory at the given path and invokes the callback for each entry.
func (fs *Unixfs) Walk(fpath string, cb WalkFun) error {
	node, err := fs.dir.Find(fs.ctx, fpath)
	if err != nil {
		return err
	}

	dir, err := ufsio.NewDirectoryFromNode(fs.ds, node)
	if err != nil {
		return err
	}

	return fs.walkPath(fpath, dir, cb)
}

func (fs *Unixfs) walkPath(fpath string, dir ufsio.Directory, cb WalkFun) error {
	links, err := dir.Links(fs.ctx)
	if err != nil {
		return err
	}

	for _, l := range links {
		subnode, err := fs.ds.Get(fs.ctx, l.Cid)
		if err != nil {
			return err
		}

		fsnode, err := ufs.ExtractFSNode(subnode)
		if err != nil {
			return err
		}

		subpath := path.Join(fpath, l.Name)
		if err := cb(subpath, fsnode); err != nil {
			return err
		}

		if !fsnode.IsDir() {
			return nil
		}

		dir, err = ufsio.NewDirectoryFromNode(fs.ds, subnode)
		if err != nil {
			return err
		}

		if err := fs.walkPath(subpath, dir, cb); err != nil {
			return err
		}
	}

	return nil
}

func (fs *Unixfs) findPath(dir ufsio.Directory, parts []string) (ipld.Node, error) {
	if len(parts) == 0 {
		return nil, errors.New("invalid file path")
	}

	if len(parts) == 1 {
		return dir.Find(fs.ctx, parts[0])
	}

	node, err := dir.Find(fs.ctx, parts[0])
	if err != nil {
		return nil, err
	}

	sub, err := ufsio.NewDirectoryFromNode(fs.ds, node)
	if err != nil {
		return nil, err
	}

	return fs.findPath(sub, parts[1:])
}

func (fs *Unixfs) removePath(dir ufsio.Directory, parts []string) error {
	if len(parts) == 0 {
		return errors.New("invalid file path")
	}

	if len(parts) == 1 {
		return dir.RemoveChild(fs.ctx, parts[0])
	}

	node, err := dir.Find(fs.ctx, parts[0])
	if err != nil {
		return err
	}

	sub, err := ufsio.NewDirectoryFromNode(fs.ds, node)
	if err != nil {
		return err
	}

	if err := fs.removePath(sub, parts[1:]); err != nil {
		return err
	}

	node, err = sub.GetNode()
	if err != nil {
		return err
	}

	if err := fs.ds.Add(fs.ctx, node); err != nil {
		return err
	}

	return dir.AddChild(fs.ctx, parts[0], node)
}

func (fs *Unixfs) writePath(dir ufsio.Directory, parts []string, file ipld.Node) error {
	if len(parts) == 0 {
		return errors.New("invalid file path")
	}

	if len(parts) == 1 {
		return dir.AddChild(fs.ctx, parts[0], file)
	}

	node, err := dir.Find(fs.ctx, parts[0])
	if err != nil && err != os.ErrNotExist {
		return err
	}

	if err == os.ErrNotExist {
		node = ufs.EmptyDirNode()
	}

	sub, err := ufsio.NewDirectoryFromNode(fs.ds, node)
	if err != nil {
		return err
	}

	if fs.writePath(sub, parts[1:], file); err != nil {
		return err
	}

	node, err = sub.GetNode()
	if err != nil {
		return err
	}

	if err := fs.ds.Add(fs.ctx, node); err != nil {
		return err
	}

	return dir.AddChild(fs.ctx, parts[0], node)
}
