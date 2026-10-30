// Package config reads server settings from environment variables.
package config

import (
	"os"
	"strconv"
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
	// CupsSpeedup divides every three cups duration, so browser tests do not
	// take minutes (RW-005). Env: RW_CUPS_FAST=1 sets 10; default 1 (real speed).
	CupsSpeedup int
	// RummyFast shortens the rummy lobby countdown for browser tests (RW-006).
	// Env: RW_RUMMY_FAST=1; default off.
	RummyFast bool
	// RummySeed fixes the rummy shuffles so browser tests know the hands
	// (RW-006). Env: RW_RUMMY_SEED=<n>; default 0 (random).
	RummySeed uint64
}

// FromEnv loads the configuration, applying defaults.
func FromEnv() Config {
	return Config{
		Addr:           getenv("RW_ADDR", ":8080"),
		AllowedOrigins: splitList(getenv("RW_ALLOWED_ORIGINS", "localhost:3000")),
		Debug:          os.Getenv("RW_DEBUG") == "1",
		CupsSpeedup:    cupsSpeedup(),
		RummyFast:      os.Getenv("RW_RUMMY_FAST") == "1",
		RummySeed:      rummySeed(),
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

// FastCupsSpeedup is the speedup RW_CUPS_FAST=1 turns on.
const FastCupsSpeedup = 10

func cupsSpeedup() int {
	if os.Getenv("RW_CUPS_FAST") == "1" {
		return FastCupsSpeedup
	}
	return 1
}

// rummySeed reads RW_RUMMY_SEED; anything that is not a positive number
// means a random seed.
func rummySeed() uint64 {
	n, err := strconv.ParseUint(os.Getenv("RW_RUMMY_SEED"), 10, 64)
	if err != nil {
		return 0
	}
	return n
}
