package main

import (
	"io"
	"os"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/pass"
	"github.com/labstack/echo"
)

func main() {
	cfg := &webpass.Config{
		Host: ":8080",
		Store: pass.NewDefaultStore(),
		OpenPGPKey: func() (io.ReadCloser, error) {
			return os.Open("private-key.gpg")
		},
	}

	e := echo.New()
	e.Logger.Fatal(webpass.Start(e, cfg))
}
