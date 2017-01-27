package webpass

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emersion/webpass/pass"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func generateKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", key), nil
}

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

type (
	authReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	authRes struct {
		Key string `json:"key"`
	}
)

type apiError struct {
	Title string `json:"title"`
}

func newAPIError(err error) *apiError {
	return &apiError{err.Error()}
}

type Server struct {
	Addr    string
	Backend Backend

	sessions map[string]User
}

func NewServer(be Backend) *Server {
	return &Server{
		Backend:  be,
		sessions: make(map[string]User),
	}
}

func (s *Server) Start(e *echo.Echo) error {
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

	api := e.Group("/pass")

	api.POST("/auth", func(c echo.Context) error {
		var req authReq
		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
			return err
		}

		u, err := s.Backend.Auth(req.Username, req.Password)
		if err == ErrInvalidCredentials {
			return echo.NewHTTPError(http.StatusUnauthorized)
		} else if err != nil {
			return err
		}

		key, err := generateKey()
		if err != nil {
			return err
		}

		s.sessions[key] = u
		return c.JSON(http.StatusOK, &authRes{
			Key: key,
		})
	})

	authAPI := api.Group("")

	authAPI.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(key string, c echo.Context) bool {
			u, ok := s.sessions[key]
			if !ok {
				return false
			}

			c.Set("user", u)
			return true
		},
	}))

	authAPI.GET("/store", func(c echo.Context) error {
		list, err := userFrom(c).List()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, list)
	})
	authAPI.GET("/store/*.gpg", func(c echo.Context) error {
		r, err := userFrom(c).Open(c.Param("*"))
		if err == pass.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		} else if err != nil {
			return err
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-encrypted", r)
	})
	authAPI.GET("/keys.gpg", func(c echo.Context) error {
		r, err := userFrom(c).OpenPGPKey()
		if err == ErrNoSuchKey {
			return echo.NewHTTPError(http.StatusNotFound)
		} else if err != nil {
			return err
		}
		defer r.Close()

		return c.Stream(http.StatusOK, "application/pgp-keys", r)
	})

	e.Static("/node_modules", "node_modules")
	e.Static("/", "public")

	return e.Start(s.Addr)
}
