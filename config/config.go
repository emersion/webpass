package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/pass"
)

type AuthFunc func(username, password string) error
type AuthCreateFunc func() (AuthFunc, error)

var auths = make(map[string]AuthCreateFunc)

type backend struct {
	auth AuthFunc
}

func (be *backend) Auth(username, password string) (webpass.User, error) {
	if err := be.auth(username, password); err != nil {
		return nil, err
	}

	u := &user{pass.NewDefaultStore()}
	return u, nil
}

type user struct {
	s *pass.Store
}

func (u *user) Store() *pass.Store {
	return u.s
}

func (u *user) OpenPGPKey() (io.ReadCloser, error) {
	return os.Open("private-key.gpg")
}

type Config struct {
	Auth string `json:"auth"`
}

func New() *Config {
	return &Config{
		Auth: "pam",
	}
}

func Open(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := new(Config)
	err = json.NewDecoder(f).Decode(cfg)
	return cfg, err
}

func (cfg *Config) Backend() (webpass.Backend, error) {
	be := new(backend)

	createAuth, ok := auths[cfg.Auth]
	if !ok {
		return nil, fmt.Errorf("Unknown auth %q", cfg.Auth)
	}
	auth, err := createAuth()
	if err != nil {
		return nil, err
	}
	be.auth = auth

	return be, nil
}
