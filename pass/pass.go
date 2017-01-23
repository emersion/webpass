package pass

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Store struct {
	Path string
}

func NewDefaultStore() *Store {
	return &Store{
		Path: defaultStorePath(),
	}
}

func defaultStorePath() string {
	if path := os.Getenv("PASSWORD_STORE_DIR"); path != "" {
		return path
	}
	return filepath.Join(os.Getenv("HOME"), ".password-store")
}

func (s *Store) List() ([]string, error) {
	var list []string

	err := filepath.Walk(s.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == s.Path {
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

		item, err := filepath.Rel(s.Path, path)
		if err != nil {
			return err
		}

		list = append(list, item)
		return nil
	})

	return list, err
}

func (s *Store) Open(item string) (io.ReadCloser, error) {
	p := filepath.Join(s.Path, item)
	if !filepath.HasPrefix(p, s.Path) {
		// Make sure the requested item is *in* the password store
		return nil, errors.New("Invalid item path")
	}
	return os.Open(p)
}
