package webpass

import (
	"errors"
	"io"

	"github.com/emersion/webpass/pass"
)

var ErrInvalidCredentials = errors.New("Invalid credentials")

type Backend interface {
	Auth(username, password string) (User, error)
}

type User interface {
	Store() *pass.Store
	OpenPGPKey() (io.ReadCloser, error)
}
