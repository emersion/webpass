package webpass

import (
	"github.com/labstack/echo"
)

type Config struct {
	Host string
}

func Start(e *echo.Echo, cfg *Config) error {
	e.Static("/node_modules", "node_modules")
	e.Static("/", "public")

	return e.Start(cfg.Host)
}
