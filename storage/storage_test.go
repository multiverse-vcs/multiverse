package storage

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/storage/test"
	"github.com/ipfs/go-merkledag/dagutils"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type StorageSuite struct {
	test.BaseStorageSuite
}

var _ = Suite(&StorageSuite{})

func (s *StorageSuite) SetUpTest(c *C) {
	ctx := context.Background()
	ds := dagutils.NewMemoryDagService()

	fs, err := unixfs.New(ctx, ds)
	if err != nil {
		c.Fatalf("failed to create unixfs: %v", err)
	}

	storer := NewStorage(fs)
	s.BaseStorageSuite = test.NewBaseStorageSuite(storer)
}
