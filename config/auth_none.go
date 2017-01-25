package config

func init() {
	auths["none"] = createAuthNone
}

func createAuthNone() (AuthFunc, error) {
	return func(username, password string) error {
		return nil
	}, nil
}
