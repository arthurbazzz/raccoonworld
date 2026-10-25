package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"raccoons-world/server/internal/adapter/httpapi"
)

func TestHealth(t *testing.T) {
	r := httpapi.NewRouter(http.NotFoundHandler())
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

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
	r := httpapi.NewRouter(ws)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ws", nil))

	if !called || rec.Code != http.StatusTeapot {
		t.Fatalf("GET /ws: called=%v code=%d, want delegated to the WebSocket handler", called, rec.Code)
	}
}
