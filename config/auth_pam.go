// +build linux darwin

package config

import (
	"errors"
	osuser "os/user"

	"github.com/emersion/webpass"
	"github.com/msteinert/pam"
)

func init() {
	auths["pam"] = createAuthPAM
}

func createAuthPAM() (AuthFunc, error) {
	u, err := osuser.Current()
	if err != nil {
		return nil, err
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
	}, nil
}
