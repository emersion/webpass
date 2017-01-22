package main

import (
	"github.com/emersion/webpass"
	"github.com/labstack/echo"
)

func main() {
	cfg := &webpass.Config{
		Host: ":8080",
	}

	e := echo.New()
	e.Logger.Fatal(webpass.Start(e, cfg))
}
