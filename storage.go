package ipfs

import (
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

// Storage is an implementation of git.Storer
type Storage struct {
	ConfigStorage
	ObjectStorage
	ShallowStorage
	IndexStorage
	ReferenceStorage
	ModuleStorage
}

// NewStorage returns a storer using a unixfs directory.
func NewStorage(fs *unixfs.Unixfs) *Storage {
	return &Storage{
		ConfigStorage:    ConfigStorage{},
		ShallowStorage:   ShallowStorage{},
		ReferenceStorage: NewReferenceStorage(fs),
		ObjectStorage:    NewObjectStorage(fs),
		ModuleStorage:    NewModuleStorage(fs),
	}
}
