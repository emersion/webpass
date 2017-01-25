package main

import (
	"os"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/config"
	"github.com/labstack/echo"
)

func main() {
	e := echo.New()

	host := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		host = ":"+port
	}

	cfg, err := config.Open("config.json")
	if os.IsNotExist(err) {
		cfg = config.New()
	} else if err != nil {
		e.Logger.Fatal(err)
	}

	be, err := cfg.Backend()
	if err != nil {
		e.Logger.Fatal(err)
	}

	s := &webpass.Server{
		Host: host,
		Backend: be,
	}

	e.Logger.Fatal(s.Start(e))
}
