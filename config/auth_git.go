package config

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/pass"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	transporthttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	transportssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func init() {
	auths["git"] = createAuthGit
}

type gitConfig struct {
	URL        string `json:"url"`
	PrivateKey string `json:"privatekey,omitempty"`
}

func createAuthGit(rawConfig json.RawMessage) (AuthFunc, error) {
	cfg := new(gitConfig)
	if err := json.Unmarshal(rawConfig, cfg); err != nil {
		return nil, err
	}

	return func(username, password string) (pass.Store, error) {
		r := git.NewMemoryRepository()

		ustr := cfg.URL
		if !strings.Contains(ustr, "://") {
			ustr = "ssh://" + ustr
		}

		u, err := url.Parse(ustr)
		if err != nil {
			return nil, err
		}

		var auth transport.AuthMethod
		switch u.Scheme {
		case "http", "https":
			auth = transporthttp.NewBasicAuth(username, password)
		case "git+ssh", "ssh", "":
			if cfg.PrivateKey != "" {
				b, err := ioutil.ReadFile(cfg.PrivateKey)
				if err != nil {
					return nil, err
				}

				block, _ := pem.Decode(b)
				if block == nil {
					return nil, fmt.Errorf("key is not PEM-encoded")
				}

				if x509.IsEncryptedPEMBlock(block) {
					block.Bytes, err = x509.DecryptPEMBlock(block, []byte(password))
					if err != nil {
						return nil, err
					}

					procType := strings.Split(block.Headers["Proc-Type"], ",")
					var newProcType []string
					for _, t := range procType {
						if t != "ENCRYPTED" {
							newProcType = append(newProcType, t)
						}
					}
					block.Headers["Proc-Type"] = strings.Join(newProcType, ",")
				}

				b = pem.EncodeToMemory(block)
				signer, err := ssh.ParsePrivateKey(b)
				if err != nil {
					return nil, err
				}

				var user string
				if u.User != nil {
					user = u.User.Username()
				}

				auth = &transportssh.PublicKeys{
					User:   user,
					Signer: signer,
				}
			} else {
				auth = &transportssh.Password{
					User: username,
					Pass: password,
				}
			}
		}

		err = r.Clone(&git.CloneOptions{
			URL:   cfg.URL,
			Auth:  auth,
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
