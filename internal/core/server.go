package core

import (
	"context"
	"io"
	"os"
	"path/filepath"

	config "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/core"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
)

type Server struct {
	Node *core.IpfsNode
	DB   *gorm.DB
}

// NewServer returns a new server.
func NewServer(ctx context.Context) (*Server, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	rpath := filepath.Join(home, ".multiverse")
	ppath := filepath.Join(rpath, "plugins")
	dpath := filepath.Join(rpath, "multiverse.db")

	plugins, err := loader.NewPluginLoader(ppath)
	if err != nil {
		return nil, err
	}

	if err := plugins.Initialize(); err != nil {
		return nil, err
	}

	if err := plugins.Inject(); err != nil {
		return nil, err
	}

	cfg, err := config.Init(io.Discard, 2048)
	if err != nil {
		return nil, err
	}

	if err := fsrepo.Init(rpath, cfg); err != nil {
		return nil, err
	}

	repo, err := fsrepo.Open(rpath)
	if err != nil {
		return nil, err
	}

	opts := &core.BuildCfg{
		Online:    true,
		Permanent: true,
		Routing:   libp2p.DHTOption,
		Repo:      repo,
	}

	node, err := core.NewNode(ctx, opts)
	if err != nil {
		return nil, err
	}

	db, err := database.Open(sqlite.Open(dpath))
	if err != nil {
		return nil, err
	}

	return &Server{
		Node: node,
		DB:   db,
	}, nil
}
