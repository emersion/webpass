package pass

import (
	"errors"
	"io"
)

var ErrNotFound = errors.New("pass: not found")

type Store interface {
	List() ([]string, error)
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
}
