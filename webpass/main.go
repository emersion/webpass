package main

import (
	"io"
	"os"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/backend/pam"
	"github.com/emersion/webpass/backend/public"
	"github.com/emersion/webpass/pass"
	"github.com/labstack/echo"
)

type backend struct {
	auth func(username, password string) error
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

type config struct {
	Auth string `json:"auth"`
}

func (c *config) backend() webpass.Backend {
	be := new(backend)

	switch c.Auth {
	case "none":
		be.auth = public.Auth
	default: // "pam"
		be.auth = pam.NewAuth()
	}

	return be
}

func main() {
	e := echo.New()

	cfg := new(config)

	host := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		host = ":"+port
	}

	s := &webpass.Server{
		Host: host,
		Backend: cfg.backend(),
	}

	e.Logger.Fatal(s.Start(e))
}
