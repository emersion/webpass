package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/pass"
)

type (
	AuthFunc       func(username, password string) (pass.Store, error)
	AuthCreateFunc func(config json.RawMessage) (AuthFunc, error)
)

var auths = make(map[string]AuthCreateFunc)

type backend struct {
	config *Config
	auth AuthFunc
}

func (be *backend) Auth(username, password string) (webpass.User, error) {
	s, err := be.auth(username, password)
	if err != nil {
		return nil, err
	}

	if s == nil {
		s = pass.NewDefaultStore()
	}

	u, ok := s.(webpass.User)
	if !ok {
		pgpConfig := be.config.PGP
		if pgpConfig == nil {
			pgpConfig = &PGPConfig{PrivateKey: "private-key.gpg"}
		}

		u = &user{
			Store: s,
			pgpConfig: pgpConfig,
		}
	}
	return u, nil
}

type user struct {
	pass.Store
	pgpConfig *PGPConfig
}

func (u *user) OpenPGPKey() (io.ReadCloser, error) {
	if u.pgpConfig == nil || u.pgpConfig.PrivateKey == "" {
		return nil, webpass.ErrNoSuchKey
	}

	f, err := os.Open(u.pgpConfig.PrivateKey)
	if os.IsNotExist(err) {
		return nil, webpass.ErrNoSuchKey
	}
	return f, err
}

type authConfig struct {
	Type string `json:"type"`
}

type Config struct {
	AuthType string `json:"-"`
	PGP *PGPConfig `json:"pgp"`

	Auth json.RawMessage `json:"auth,omitempty"`
}

type PGPConfig struct{
	PrivateKey string `json:"privatekey,omitempty"`
}

func New() *Config {
	return &Config{
		AuthType: "pam",
	}
}

func Open(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := new(Config)
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	auth := new(authConfig)
	if err := json.Unmarshal(cfg.Auth, auth); err != nil {
		return nil, err
	}
	cfg.AuthType = auth.Type

	return cfg, err
}

func (cfg *Config) Backend() (webpass.Backend, error) {
	be := &backend{config: cfg}

	createAuth, ok := auths[cfg.AuthType]
	if !ok {
		return nil, fmt.Errorf("Unknown auth %q", cfg.Auth)
	}
	auth, err := createAuth(cfg.Auth)
	if err != nil {
		return nil, err
	}
	be.auth = auth

	return be, nil
}
