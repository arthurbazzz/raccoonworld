package features

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/cucumber/godog"

	"raccoons-world/server/internal/app"
	"raccoons-world/server/internal/infra/clock"
)

// TestSmokeOverWebSocket runs the @smoke scenarios against the real server
// (app.New, the same wiring as cmd/server) over real WebSocket connections.
// It proves transport, protocol and use cases work together end to end.
func TestSmokeOverWebSocket(t *testing.T) {
	tags := "@smoke && ~@ui"
	if extra := os.Getenv("GODOG_TAGS"); extra != "" {
		tags += " && " + extra
	}
	suite := godog.TestSuite{
		Name:                "rw-001-smoke",
		ScenarioInitializer: initializeSmokeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    featurePaths,
			Tags:     tags,
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("smoke scenarios failed")
	}
}

// waitTimeout bounds how long a step waits for a server message.
const waitTimeout = 2 * time.Second

// wireMessage is a protocol v1 message as seen by a client.
type wireMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// wsClient is one browser-like connection that records what it receives.
type wsClient struct {
	conn *websocket.Conn

	mu       sync.Mutex
	received []wireMessage
}

func (c *wsClient) send(typ string, data any) error {
	raw, err := json.Marshal(map[string]any{"type": typ, "v": 1, "data": data})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	return c.conn.Write(ctx, websocket.MessageText, raw)
}

func (c *wsClient) readLoop() {
	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			return
		}
		var m wireMessage
		if json.Unmarshal(data, &m) == nil {
			c.mu.Lock()
			c.received = append(c.received, m)
			c.mu.Unlock()
		}
	}
}

// waitFor polls received messages until one of type typ satisfies match.
func (c *wsClient) waitFor(typ string, match func(json.RawMessage) bool) error {
	return c.waitForWithin(waitTimeout, typ, match)
}

// waitForWithin is waitFor with its own time limit.
func (c *wsClient) waitForWithin(limit time.Duration, typ string, match func(json.RawMessage) bool) error {
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		for _, m := range c.received {
			if m.Type == typ && match(m.Data) {
				c.mu.Unlock()
				return nil
			}
		}
		c.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("no %s message matched within %v", typ, limit)
}

type smokeWorld struct {
	srv     *httptest.Server
	stop    context.CancelFunc
	roomEnd chan struct{}

	clients     map[string]*wsClient // nickname -> joined connection
	ids         map[string]string    // nickname -> player id from room.joined
	lastAttempt *wsClient
	all         []*wsClient
}

func initializeSmokeScenario(sc *godog.ScenarioContext) {
	w := &smokeWorld{clients: map[string]*wsClient{}, ids: map[string]string{}}
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.close()
		return ctx, nil
	})

	sc.Step(`^que a sala "([^"]*)" está aberta$`, w.openRoom)
	sc.Step(`^o visitante entra com o apelido "([^"]*)"$`, w.join)
	sc.Step(`^que "([^"]*)" está na Praça$`, w.join)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão na Praça$`, w.joinBoth)
	sc.Step(`^outro visitante tenta entrar com o apelido "([^"]*)"$`, w.attemptJoin)
	sc.Step(`^"([^"]*)" aparece na Praça$`, w.appearsInRoom)
	sc.Step(`^"([^"]*)" recebe a lista dos jogadores já presentes$`, w.receivesPlayerList)
	sc.Step(`^a entrada é recusada com o motivo "([^"]*)"$`, w.joinRejectedWith)
	sc.Step(`^"([^"]*)" envia a mensagem "([^"]*)"$`, w.say)
	sc.Step(`^"([^"]*)" vê o balão "([^"]*)" sobre "([^"]*)"$`, w.seesBalloon)
	w.registerCupsSmokeSteps(sc)
}

func (w *smokeWorld) openRoom(name string) error {
	if _, ok := roomIDs[name]; !ok {
		return fmt.Errorf("unknown room %q", name)
	}
	// Fast three cups (RW_CUPS_FAST), so a round takes a second or two.
	a, err := app.New(app.Options{Clock: clock.System{}, CupsSpeedup: 10})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.stop = cancel
	w.roomEnd = make(chan struct{})
	go func() {
		defer close(w.roomEnd)
		a.RunRoom(ctx)
	}()
	w.srv = httptest.NewServer(a.Handler)
	return nil
}

func (w *smokeWorld) close() {
	for _, c := range w.all {
		_ = c.conn.CloseNow()
	}
	if w.srv != nil {
		w.srv.Close()
	}
	if w.stop != nil {
		w.stop()
		<-w.roomEnd
	}
}

func (w *smokeWorld) connect() (*wsClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	url := "ws" + strings.TrimPrefix(w.srv.URL, "http") + "/ws"
	conn, resp, err := websocket.Dial(ctx, url, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	c := &wsClient{conn: conn}
	go c.readLoop()
	w.all = append(w.all, c)
	return c, nil
}

func (w *smokeWorld) attemptJoin(nickname string) error {
	c, err := w.connect()
	if err != nil {
		return err
	}
	w.lastAttempt = c
	return c.send("room.join", map[string]string{"nickname": nickname})
}

// join connects, joins and waits for room.joined to learn the player id.
func (w *smokeWorld) join(nickname string) error {
	if err := w.attemptJoin(nickname); err != nil {
		return err
	}
	c := w.lastAttempt
	var id string
	err := c.waitFor("room.joined", func(data json.RawMessage) bool {
		var d struct {
			You string `json:"you"`
		}
		if json.Unmarshal(data, &d) != nil || d.You == "" {
			return false
		}
		id = d.You
		return true
	})
	if err != nil {
		return fmt.Errorf("%q did not join: %w", nickname, err)
	}
	w.clients[nickname], w.ids[nickname] = c, id
	return nil
}

func (w *smokeWorld) joinBoth(a, b string) error {
	if err := w.join(a); err != nil {
		return err
	}
	return w.join(b)
}

type joinedData struct {
	You     string `json:"you"`
	Players []struct {
		ID       string `json:"id"`
		Nickname string `json:"nickname"`
	} `json:"players"`
}

func (w *smokeWorld) client(nickname string) (*wsClient, error) {
	c, ok := w.clients[nickname]
	if !ok {
		return nil, fmt.Errorf("%q never joined", nickname)
	}
	return c, nil
}

func (w *smokeWorld) appearsInRoom(nickname string) error {
	c, err := w.client(nickname)
	if err != nil {
		return err
	}
	// The room announces every arrival, the newcomer included.
	return c.waitFor("player.joined", func(data json.RawMessage) bool {
		var d struct {
			Nickname string `json:"nickname"`
		}
		return json.Unmarshal(data, &d) == nil && d.Nickname == nickname
	})
}

func (w *smokeWorld) receivesPlayerList(nickname string) error {
	c, err := w.client(nickname)
	if err != nil {
		return err
	}
	return c.waitFor("room.joined", func(data json.RawMessage) bool {
		var d joinedData
		if json.Unmarshal(data, &d) != nil || len(d.Players) != len(w.clients) {
			return false
		}
		for _, p := range d.Players {
			if w.ids[p.Nickname] != p.ID {
				return false
			}
		}
		return true
	})
}

func (w *smokeWorld) joinRejectedWith(reason string) error {
	want, ok := rejectReasons[reason]
	if !ok {
		return fmt.Errorf("unknown reason %q", reason)
	}
	if w.lastAttempt == nil {
		return fmt.Errorf("nobody tried to join")
	}
	return w.lastAttempt.waitFor("room.join_rejected", func(data json.RawMessage) bool {
		var d struct {
			Reason string `json:"reason"`
		}
		return json.Unmarshal(data, &d) == nil && d.Reason == string(want)
	})
}

func (w *smokeWorld) say(nickname, text string) error {
	c, err := w.client(nickname)
	if err != nil {
		return err
	}
	return c.send("chat.send", map[string]string{"text": text})
}

func (w *smokeWorld) seesBalloon(observer, text, sender string) error {
	c, err := w.client(observer)
	if err != nil {
		return err
	}
	senderID := w.ids[sender]
	return c.waitFor("chat.message", func(data json.RawMessage) bool {
		var d struct {
			From string `json:"from"`
			Text string `json:"text"`
		}
		return json.Unmarshal(data, &d) == nil && d.From == senderID && d.Text == text
	})
}
