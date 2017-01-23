package webpass

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func userFrom(c echo.Context) User {
	v := c.Get("user")
	if v == nil {
		return nil
	}
	u, ok := v.(User)
	if !ok {
		return nil
	}
	return u
}

type Server struct {
	Host string
	Backend Backend
}

func (s *Server) Start(e *echo.Echo) error {
	g := e.Group("/pass")

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 'unsafe-eval' is needed by vue, 'unsafe-inline' is needed by vue-material
			const csp = "default-src 'self' 'unsafe-eval' 'unsafe-inline';" +
				"object-src 'none';" +
				"frame-ancestors 'none';" +
				"form-action 'none';" +
				"block-all-mixed-content"

			h := c.Response().Header()
			h.Set("X-Frame-Options", "DENY")
			h.Set("Content-Security-Policy", csp)

			return next(c)
		}
	})

	g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) bool {
		u, err := s.Backend.Auth(username, password)
		if err != nil {
			if err != ErrInvalidCredentials {
				e.Logger.Error("Cannot authenticate user:", err)
			}
			return false
		}

		c.Set("user", u)
		return true
	}))

	g.GET("/store", func(c echo.Context) error {
		list, err := userFrom(c).Store().List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, newAPIError(err))
		}
		return c.JSON(http.StatusOK, list)
	})
	g.GET("/store/*.gpg", func(c echo.Context) error {
		r, err := userFrom(c).Store().Open(c.Param("*"))
		if err != nil {
			return c.JSON(http.StatusNotFound, newAPIError(err))
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-encrypted", r)
	})
	g.GET("/keys.gpg", func(c echo.Context) error {
		r, err := userFrom(c).OpenPGPKey()
		if err != nil {
			return c.JSON(http.StatusNotFound, newAPIError(err))
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-keys", r)
	})

	e.Static("/node_modules", "node_modules")
	e.Static("/", "public")

	return e.Start(s.Host)
}

type apiError struct {
	Title string `json:"title"`
}

func newAPIError(err error) *apiError {
	return &apiError{err.Error()}
}
