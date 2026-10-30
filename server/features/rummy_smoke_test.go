package features

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"
)

// rummyStateData is the rummy.state message as a client reads it.
type rummyStateData struct {
	Hand []int `json:"hand"`
	Pool int   `json:"pool"`
}

// Smoke steps for the rummy table, over real WebSocket connections and a
// short lobby (RummyFast).
func (w *smokeWorld) registerRummySmokeSteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a mesa de peças está livre$`, func() error { return nil }) // a fresh server per scenario
	sc.Step(`^começa uma partida entre "([^"]*)" e "([^"]*)"$`, w.startRummyMatch)
	sc.Step(`^cada um tem (\d+) peças na mão$`, w.everyRummyHandHas)
	sc.Step(`^o monte tem (\d+) peças$`, w.rummyPoolHas)
}

// walkToTable holds "down" from the spawn point until the raccoon is right
// above the table.
func (w *smokeWorld) walkToTable(nickname string) error {
	c, err := w.client(nickname)
	if err != nil {
		return err
	}
	if err := c.send("player.input", map[string]bool{"down": true}); err != nil {
		return err
	}
	deadline := time.Now().Add(walkLimit)
	for {
		if _, y, ok := w.position(nickname); ok && y >= 436 {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%q did not reach the table", nickname)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return c.send("player.input", map[string]bool{})
}

func (w *smokeWorld) startRummyMatch(a, b string) error {
	if err := w.joinBoth(a, b); err != nil {
		return err
	}
	for _, n := range []string{a, b} {
		if err := w.walkToTable(n); err != nil {
			return err
		}
	}
	ca, _ := w.client(a)
	if err := ca.send("rummy.create", map[string]string{"mode": "free"}); err != nil {
		return err
	}
	var roomID string
	if err := ca.waitFor("rummy.rooms", func(d json.RawMessage) bool {
		var r struct {
			Rooms []struct {
				ID      string   `json:"id"`
				Players []string `json:"players"`
			} `json:"rooms"`
		}
		if json.Unmarshal(d, &r) != nil {
			return false
		}
		for _, room := range r.Rooms {
			if slices.Contains(room.Players, w.ids[a]) {
				roomID = room.ID
				return true
			}
		}
		return false
	}); err != nil {
		return fmt.Errorf("%q's room never showed up: %w", a, err)
	}
	cb, _ := w.client(b)
	if err := cb.send("rummy.join", map[string]string{"room": roomID}); err != nil {
		return err
	}
	for _, c := range []*wsClient{ca, cb} {
		if err := c.waitForWithin(5*time.Second, "rummy.state", func(json.RawMessage) bool { return true }); err != nil {
			return fmt.Errorf("the match did not start: %w", err)
		}
	}
	return nil
}

func (w *smokeWorld) rummyState(nickname string) (rummyStateData, error) {
	c, err := w.client(nickname)
	if err != nil {
		return rummyStateData{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.received) - 1; i >= 0; i-- {
		if m := c.received[i]; m.Type == "rummy.state" {
			var s rummyStateData
			return s, json.Unmarshal(m.Data, &s)
		}
	}
	return rummyStateData{}, fmt.Errorf("%q got no rummy.state", nickname)
}

func (w *smokeWorld) everyRummyHandHas(n int) error {
	for nick := range w.clients {
		s, err := w.rummyState(nick)
		if err != nil {
			return err
		}
		if len(s.Hand) != n {
			return fmt.Errorf("%q has %d tiles, want %d", nick, len(s.Hand), n)
		}
	}
	return nil
}

func (w *smokeWorld) rummyPoolHas(n int) error {
	for nick := range w.clients {
		s, err := w.rummyState(nick)
		if err != nil {
			return err
		}
		if s.Pool != n {
			return fmt.Errorf("%q sees %d in the pool, want %d", nick, s.Pool, n)
		}
	}
	return nil
}
