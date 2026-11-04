package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	cases := []struct {
		name   string
		status int
		ok     bool
	}{
		{"healthy", http.StatusOK, true},
		{"unhealthy", http.StatusServiceUnavailable, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/health" {
					t.Errorf("path = %q, want /health", r.URL.Path)
				}
				w.WriteHeader(c.status)
			}))
			defer srv.Close()

			err := healthcheck(t.Context(), strings.TrimPrefix(srv.URL, "http://"))
			if (err == nil) != c.ok {
				t.Fatalf("healthcheck = %v, want ok=%v", err, c.ok)
			}
		})
	}
}

func TestHealthcheckServerDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := strings.TrimPrefix(srv.URL, "http://")
	srv.Close()

	if err := healthcheck(t.Context(), addr); err == nil {
		t.Fatal("healthcheck with no server = nil, want an error")
	}
}

func TestHealthcheckHost(t *testing.T) {
	cases := map[string]string{
		":8080":          "127.0.0.1:8080",
		"0.0.0.0:9000":   "127.0.0.1:9000",
		"127.0.0.1:8080": "127.0.0.1:8080",
	}
	for addr, want := range cases {
		if got := healthcheckHost(addr); got != want {
			t.Errorf("healthcheckHost(%q) = %q, want %q", addr, got, want)
		}
	}
}
