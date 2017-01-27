package config

import (
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/pass"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func init() {
	auths["git"] = createAuthGit
}

type gitConfig struct {
	URL string `json:"url"`
}

func createAuthGit(rawConfig json.RawMessage) (AuthFunc, error) {
	cfg := new(gitConfig)
	if err := json.Unmarshal(rawConfig, cfg); err != nil {
		return nil, err
	}

	return func(username, password string) (pass.Store, error) {
		r := git.NewMemoryRepository()

		var auth transport.AuthMethod = http.NewBasicAuth(username, password)
		err := r.Clone(&git.CloneOptions{
			URL: cfg.URL,
			Auth: auth,
			Depth: 1,
		})
		if err == transport.ErrAuthorizationRequired {
			return nil, webpass.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}

		return &gitStore{r}, nil
	}, nil
}

type gitStore struct {
	r *git.Repository
}

func (s *gitStore) head() (*object.Commit, error) {
	ref, err := s.r.Head()
	if err != nil {
		return nil, err
	}

	return s.r.Commit(ref.Hash())
}

func (s *gitStore) List() ([]string, error) {
	commit, err := s.head()
	if err != nil {
		return nil, err
	}

	files, err := commit.Files()
	if err != nil {
		return nil, err
	}

	var list []string
	err = files.ForEach(func(f *object.File) error {
		if f.Mode.IsDir() {
			return nil
		}
		if name := filepath.Base(f.Name); len(name) > 0 && name[0] == '.' {
			return nil
		}

		list = append(list, f.Name)
		return nil
	})
	return list, err
}

func (s *gitStore) Open(name string) (io.ReadCloser, error) {
	commit, err := s.head()
	if err != nil {
		return nil, err
	}

	f, err := commit.File(name)
	if err != nil {
		return nil, err
	}

	return f.Reader()
}
