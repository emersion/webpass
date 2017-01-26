package pass

import (
	"io"
)

type Store interface {
	List() ([]string, error)
	Open(name string) (io.ReadCloser, error)
}
