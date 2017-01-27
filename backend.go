package webpass

import (
	"errors"
	"io"

	"github.com/emersion/webpass/pass"
)

var (
	ErrInvalidCredentials = errors.New("webpass: invalid credentials")
	ErrNoSuchKey = errors.New("webpass: no such key")
)

type Backend interface {
	Auth(username, password string) (User, error)
}

type User interface {
	pass.Store

	OpenPGPKey() (io.ReadCloser, error)
}
