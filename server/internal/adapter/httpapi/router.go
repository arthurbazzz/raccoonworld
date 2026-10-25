// Package httpapi exposes the HTTP routes with Gin. Gin stays in this
// package: handlers below it are plain net/http.
package httpapi

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// Config is what the router serves.
type Config struct {
	// WebSocket handles GET /ws.
	WebSocket http.Handler
	// Maps are tile maps by name, served as JSON at GET /maps/:name.
	Maps map[string][]byte
	// AllowedOrigins are host patterns (as in the WebSocket handler) that
	// may read /maps cross-origin, e.g. the Next.js dev server.
	AllowedOrigins []string
}

// NewRouter builds the routes:
//
//	GET /health      liveness probe
//	GET /ws          WebSocket endpoint (protocol v1)
//	GET /maps/:name  tile map JSON, the single source for server and client
func NewRouter(cfg Config) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ws", gin.WrapH(cfg.WebSocket))
	r.GET("/maps/:name", func(c *gin.Context) {
		allowCORS(c, cfg.AllowedOrigins)
		data, ok := cfg.Maps[c.Param("name")]
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	})
	return r
}

// allowCORS echoes the request origin when its host matches a pattern.
func allowCORS(c *gin.Context, patterns []string) {
	c.Header("Vary", "Origin")
	origin := c.GetHeader("Origin")
	if origin == "" {
		return
	}
	u, err := url.Parse(origin)
	if err != nil {
		return
	}
	host := strings.ToLower(u.Host)
	for _, p := range patterns {
		if ok, _ := path.Match(strings.ToLower(p), host); ok {
			c.Header("Access-Control-Allow-Origin", origin)
			return
		}
	}
}
