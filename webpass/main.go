package main

import (
	"errors"
	"io"
	"os"
	osuser "os/user"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/pass"
	"github.com/labstack/echo"
	"github.com/msteinert/pam"
)

type backend struct {
	username string
}

func (be *backend) Auth(username, password string) (webpass.User, error) {
	if username == "" {
		username = be.username
	}
	if username != be.username || password == "" {
		return nil, webpass.ErrInvalidCredentials
	}

	t, err := pam.StartFunc("", username, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", errors.New("Unrecognized PAM message style")
	})
	if err != nil {
		return nil, err
	}

	if err := t.Authenticate(0); err != nil {
		return nil, webpass.ErrInvalidCredentials
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

func main() {
	e := echo.New()

	u, err := osuser.Current()
	if err != nil {
		e.Logger.Fatal(err)
	}

	be := &backend{username: u.Username}

	host := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		host = ":"+port
	}

	s := &webpass.Server{
		Host: host,
		Backend: be,
	}

	e.Logger.Fatal(s.Start(e))
}
