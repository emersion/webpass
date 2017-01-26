package config

import (
	"encoding/json"

	"github.com/emersion/webpass"
)

func init() {
	auths["none"] = createAuthNone
}

func createAuthNone(json.RawMessage) (AuthFunc, error) {
	return func(username, password string) (webpass.User, error) {
		return nil, nil
	}, nil
}
