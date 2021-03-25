package ipfs

import (
	"github.com/go-git/go-git/v5/plumbing/format/index"
)

type IndexStorage struct {
	index *index.Index
}

func (c *IndexStorage) SetIndex(idx *index.Index) error {
	c.index = idx
	return nil
}

func (c *IndexStorage) Index() (*index.Index, error) {
	if c.index == nil {
		c.index = &index.Index{Version: 2}
	}

	return c.index, nil
}
