// Package httpapi exposes the HTTP routes with Gin. Gin stays in this
// package: handlers below it are plain net/http.
package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the routes:
//
//	GET /health  liveness probe
//	GET /ws      WebSocket endpoint (protocol v1)
func NewRouter(wsHandler http.Handler) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ws", gin.WrapH(wsHandler))
	return r
}
