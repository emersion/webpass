package webpass

import (
	"net/http"
	"os"

	"github.com/emersion/webpass/pass"
	"github.com/labstack/echo"
)

type Config struct {
	Host string
}

func Start(e *echo.Echo, cfg *Config) error {
	s := pass.NewDefaultStore()

	e.GET("/pass/store", func(c echo.Context) error {
		list, err := s.List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, newAPIError(err))
		}
		return c.JSON(http.StatusOK, list)
	})
	e.GET("/pass/store/*.gpg", func(c echo.Context) error {
		r, err := s.Open(c.Param("*"))
		if err != nil {
			return c.JSON(http.StatusNotFound, newAPIError(err))
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-encrypted", r)
	})
	e.GET("/pass/keys.gpg", func(c echo.Context) error {
		f, err := os.Open("private-key.gpg")
		if err != nil {
			return c.JSON(http.StatusNotFound, newAPIError(err))
		}
		defer f.Close()

		return c.Stream(http.StatusOK, "application/pgp-keys", f)
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
