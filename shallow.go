package ipfs

import (
	"github.com/go-git/go-git/v5/plumbing"
)

type ShallowStorage []plumbing.Hash

func (s *ShallowStorage) SetShallow(commits []plumbing.Hash) error {
	*s = commits
	return nil
}

func (s ShallowStorage) Shallow() ([]plumbing.Hash, error) {
	return s, nil
}
