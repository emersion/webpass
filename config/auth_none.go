package config

import (
	"encoding/json"

	"github.com/emersion/webpass/pass"
)

func init() {
	auths["none"] = createAuthNone
}

func createAuthNone(json.RawMessage) (AuthFunc, error) {
	return func(username, password string) (pass.Store, error) {
		return nil, nil
	}, nil
}
