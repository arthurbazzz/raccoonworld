package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"raccoons-world/server/internal/adapter/httpapi"
)

func newRouter(ws http.Handler) http.Handler {
	return httpapi.NewRouter(httpapi.Config{
		WebSocket:      ws,
		Maps:           map[string][]byte{"praca": []byte(`{"name":"praca"}`)},
		AllowedOrigins: []string{"localhost:3000"},
	})
}

func get(h http.Handler, path, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := get(newRouter(http.NotFoundHandler()), "/health", "")
	if rec.Code != http.StatusOK || rec.Body.String() != `{"status":"ok"}` {
		t.Fatalf("GET /health = %d %s", rec.Code, rec.Body.String())
	}
}

func TestWebSocketRouteDelegates(t *testing.T) {
	called := false
	ws := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	rec := get(newRouter(ws), "/ws", "")
	if !called || rec.Code != http.StatusTeapot {
		t.Fatalf("GET /ws: called=%v code=%d, want delegated to the WebSocket handler", called, rec.Code)
	}
}

func TestMapServesJSON(t *testing.T) {
	rec := get(newRouter(http.NotFoundHandler()), "/maps/praca", "")
	if rec.Code != http.StatusOK || rec.Body.String() != `{"name":"praca"}` {
		t.Fatalf("GET /maps/praca = %d %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestUnknownMapIs404(t *testing.T) {
	if rec := get(newRouter(http.NotFoundHandler()), "/maps/nope", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET /maps/nope = %d, want 404", rec.Code)
	}
}

func TestMapCORS(t *testing.T) {
	r := newRouter(http.NotFoundHandler())

	rec := get(r, "/maps/praca", "http://localhost:3000")
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("allowed origin: Access-Control-Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q, want Origin", got)
	}

	rec = get(r, "/maps/praca", "http://evil.example")
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("foreign origin: Access-Control-Allow-Origin = %q, want none", got)
	}
}
