package deploy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/cucumber/godog"
)

const requestTimeout = 10 * time.Second

// registerHTTPSteps binds the steps that talk to the stack through the proxy:
// page, map, WebSocket and origin check.
func (w *stackWorld) registerHTTPSteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a stack de produção está no ar$`, w.stackIsUp)
	sc.Step(`^o navegador abre "([^"]*)"$`, w.get)
	sc.Step(`^alguém pede "([^"]*)"$`, w.get)
	sc.Step(`^a Praça carrega$`, w.pageLoads)
	sc.Step(`^recebe o mapa da Praça$`, w.receivesPracaMap)
	sc.Step(`^"([^"]*)" abre o jogo pelo endereço público$`, w.joinsThroughProxy)
	sc.Step(`^"([^"]*)" entra na Praça$`, w.isInPraca)
	sc.Step(`^um navegador em outra origem tenta abrir o WebSocket$`, w.dialFromForeignOrigin)
	sc.Step(`^a conexão é recusada$`, w.connectionRefused)
}

func (w *stackWorld) stackIsUp() error {
	if err := w.get("/health"); err != nil {
		return err
	}
	if w.status != http.StatusOK {
		return fmt.Errorf("/health status = %d, want 200", w.status)
	}
	return nil
}

// get requests path through the proxy and keeps the response.
func (w *stackWorld) get(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	w.body, err = io.ReadAll(resp.Body)
	w.status = resp.StatusCode
	return err
}

func (w *stackWorld) pageLoads() error {
	if w.status != http.StatusOK {
		return fmt.Errorf("page status = %d, want 200", w.status)
	}
	if !strings.Contains(string(w.body), "<title>Raccoons World") {
		return errors.New("page is not the Next.js game page")
	}
	return nil
}

func (w *stackWorld) receivesPracaMap() error {
	if w.status != http.StatusOK {
		return fmt.Errorf("map status = %d, want 200", w.status)
	}
	var m struct {
		Name string   `json:"name"`
		Rows []string `json:"rows"`
	}
	if err := json.Unmarshal(w.body, &m); err != nil {
		return fmt.Errorf("map is not JSON: %w", err)
	}
	if m.Name != "praca" || len(m.Rows) == 0 {
		return fmt.Errorf("map name = %q with %d rows, want the Praça", m.Name, len(m.Rows))
	}
	return nil
}

// wsURL is the public WebSocket address, as the browser builds it.
func (w *stackWorld) wsURL() string {
	return "ws" + strings.TrimPrefix(w.baseURL, "http") + "/ws"
}

// dial opens the WebSocket with the given Origin and returns the handshake status.
func (w *stackWorld) dial(origin string) (*websocket.Conn, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	conn, resp, err := websocket.Dial(ctx, w.wsURL(), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": {origin}},
	})
	status := 0
	if resp != nil {
		status = resp.StatusCode
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}
	return conn, status, err
}

// joinsThroughProxy opens the WebSocket from the public origin and asks to join.
func (w *stackWorld) joinsThroughProxy(nickname string) error {
	conn, _, err := w.dial(w.baseURL)
	if err != nil {
		return fmt.Errorf("WebSocket through the proxy: %w", err)
	}
	w.conn = conn
	raw, err := json.Marshal(map[string]any{
		"type": "room.join", "v": 1, "data": map[string]string{"nickname": nickname},
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	return conn.Write(ctx, websocket.MessageText, raw)
}

// isInPraca waits for room.joined; other messages (snapshots, chat) are skipped.
func (w *stackWorld) isInPraca(nickname string) error {
	if w.conn == nil {
		return errors.New("no WebSocket open")
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	for {
		_, raw, err := w.conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("%q did not join the Praça: %w", nickname, err)
		}
		var msg struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(raw, &msg) == nil && msg.Type == "room.joined" {
			return nil
		}
	}
}

func (w *stackWorld) dialFromForeignOrigin() error {
	conn, status, err := w.dial("http://evil.example")
	if err == nil {
		_ = conn.CloseNow()
		return errors.New("WebSocket from a foreign origin was accepted")
	}
	w.status = status
	return nil
}

func (w *stackWorld) connectionRefused() error {
	if w.status != http.StatusForbidden {
		return fmt.Errorf("foreign origin status = %d, want 403", w.status)
	}
	return nil
}
