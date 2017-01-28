package pass

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type diskStore struct {
	path string
}

func NewStore(path string) Store {
	return &diskStore{path}
}

func NewDefaultStore() Store {
	return NewStore(defaultStorePath())
}

func defaultStorePath() string {
	if path := os.Getenv("PASSWORD_STORE_DIR"); path != "" {
		return path
	}
	return filepath.Join(os.Getenv("HOME"), ".password-store")
}

func (s *diskStore) List() ([]string, error) {
	var list []string

	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == s.path {
			return nil
		}
		if name := info.Name(); len(name) > 0 && name[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}

		item, err := filepath.Rel(s.path, path)
		if err != nil {
			return err
		}

		list = append(list, item)
		return nil
	})

	return list, err
}

func (s *diskStore) itemPath(item string) (string, error) {
	p := filepath.Join(s.path, item)
	if !filepath.HasPrefix(p, s.path) {
		// Make sure the requested item is *in* the password store
		return "", errors.New("invalid item path")
	}
	return p, nil
}

func (s *diskStore) Open(item string) (io.ReadCloser, error) {
	p, err := s.itemPath(item)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return f, err
}

func (s *diskStore) Create(item string) (io.WriteCloser, error) {
	p, err := s.itemPath(item)
	if err != nil {
		return nil, err
	}

	return os.Create(p)
}
