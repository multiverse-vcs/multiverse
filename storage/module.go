package storage

import (
	"github.com/go-git/go-git/v5/storage"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

type ModuleStorage struct {
	fs      *unixfs.Unixfs
	modules map[string]*Storage
}

func NewModuleStorage(fs *unixfs.Unixfs) ModuleStorage {
	return ModuleStorage{
		fs:      fs,
		modules: make(map[string]*Storage),
	}
}

func (s *ModuleStorage) Module(name string) (storage.Storer, error) {
	if m, ok := s.modules[name]; ok {
		return m, nil
	}

	fs, err := s.fs.Module()
	if err != nil {
		return nil, err
	}

	module := NewStorage(fs)
	s.modules[name] = module
	return module, nil
}
