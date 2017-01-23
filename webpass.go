package webpass

import (
	"io"
	"net/http"

	"github.com/emersion/webpass/pass"
	"github.com/labstack/echo"
)

type Config struct {
	Host string
	Store *pass.Store
	OpenPGPKey func() (io.ReadCloser, error)
}

func Start(e *echo.Echo, cfg *Config) error {
	e.GET("/pass/store", func(c echo.Context) error {
		list, err := cfg.Store.List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, newAPIError(err))
		}
		return c.JSON(http.StatusOK, list)
	})
	e.GET("/pass/store/*.gpg", func(c echo.Context) error {
		r, err := cfg.Store.Open(c.Param("*"))
		if err != nil {
			return c.JSON(http.StatusNotFound, newAPIError(err))
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-encrypted", r)
	})
	e.GET("/pass/keys.gpg", func(c echo.Context) error {
		r, err := cfg.OpenPGPKey()
		if err != nil {
			return c.JSON(http.StatusNotFound, newAPIError(err))
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-keys", r)
	})

	e.Static("/node_modules", "node_modules")
	e.Static("/", "public")

	return e.Start(cfg.Host)
}

type apiError struct {
	Title string `json:"title"`
}

func newAPIError(err error) *apiError {
	return &apiError{err.Error()}
}
