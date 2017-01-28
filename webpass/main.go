package main

import (
	"flag"
	"os"

	"github.com/emersion/webpass"
	"github.com/emersion/webpass/config"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var configPath = flag.String("config", "config.json", "path to config file")

func main() {
	flag.Parse()

	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	cfg, err := config.Open(*configPath)
	if os.IsNotExist(err) {
		cfg = config.New()
	} else if err != nil {
		e.Logger.Fatal(err)
	}

	be, err := cfg.Backend()
	if err != nil {
		e.Logger.Fatal(err)
	}

	s := webpass.NewServer(be)
	s.Addr = addr

	e.Logger.Fatal(s.Start(e))
}
