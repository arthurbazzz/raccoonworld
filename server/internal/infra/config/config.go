// Package config reads server settings from environment variables.
package config

import (
	"os"
	"strings"
)

// Config holds the server settings.
type Config struct {
	// Addr is the HTTP listen address. Env: RW_ADDR (default ":8080").
	Addr string
	// AllowedOrigins are browser origins allowed to open the WebSocket,
	// as host patterns. Env: RW_ALLOWED_ORIGINS, comma separated
	// (default "localhost:3000", the Next.js dev server).
	AllowedOrigins []string
	// Debug enables debug logs. Env: RW_DEBUG=1.
	Debug bool
}

// FromEnv loads the configuration, applying defaults.
func FromEnv() Config {
	return Config{
		Addr:           getenv("RW_ADDR", ":8080"),
		AllowedOrigins: splitList(getenv("RW_ALLOWED_ORIGINS", "localhost:3000")),
		Debug:          os.Getenv("RW_DEBUG") == "1",
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
