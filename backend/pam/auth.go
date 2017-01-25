package pam

import (
	"errors"
	"os/user"

	"github.com/emersion/webpass"
	"github.com/msteinert/pam"
)

func NewAuth() func(username, password string) error {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	requiredUsername := u.Username

	return func(username, password string) error {
		if username == "" {
			username = requiredUsername
		}
		if username != requiredUsername || password == "" {
			return webpass.ErrInvalidCredentials
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
			return err
		}

		if err := t.Authenticate(0); err != nil {
			return webpass.ErrInvalidCredentials
		}

		return nil
	}
}
