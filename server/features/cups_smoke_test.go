package features

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/cups"
)

// cupsRoundData is the cups.round message as a client reads it.
type cupsRoundData struct {
	Round  int      `json:"round"`
	Nut    int      `json:"nut"`
	ShowMs int64    `json:"showMs"`
	Swaps  [][2]int `json:"swaps"`
	SwapMs int64    `json:"swapMs"`
}

// nut follows the swaps, as a player watching the cups does.
func (r cupsRoundData) nut() int {
	swaps := make([]cups.Swap, len(r.Swaps))
	for i, s := range r.Swaps {
		swaps[i] = cups.Swap(s)
	}
	return cups.FinalSlot(r.Nut, swaps)
}

type cupsRevealData struct {
	Picks []struct {
		ID  string `json:"id"`
		Hit bool   `json:"hit"`
	} `json:"picks"`
}

// walkLimit bounds a walk to the booth at the real speed (~2 s).
const walkLimit = 10 * time.Second

// Smoke steps for the three cups, over real WebSocket connections and a
// fast booth (CupsSpeedup 10).
func (w *smokeWorld) registerCupsSmokeSteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a barraca dos três copos está livre$`, w.boothIsFree)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão numa rodada com o embaralhamento terminado$`, w.bothChoosing)
	sc.Step(`^"([^"]*)" escolhe o copo onde está a noz$`, w.picksNut)
	sc.Step(`^"([^"]*)" escolhe um copo sem a noz$`, w.picksEmpty)
	sc.Step(`^"([^"]*)" vê "Acertou!"$`, func(n string) error { return w.seesResult(n, true) })
	sc.Step(`^"([^"]*)" vê "Errou!"$`, func(n string) error { return w.seesResult(n, false) })
	sc.Step(`^os dois veem que só "([^"]*)" acertou a rodada$`, w.everyoneSeesOnlyHit)
}

func (w *smokeWorld) boothIsFree() error { return nil } // a fresh server per scenario

// position returns the last position the server sent for nickname.
func (w *smokeWorld) position(nickname string) (x, y float64, ok bool) {
	c, id := w.clients[nickname], w.ids[nickname]
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range c.received {
		var d struct {
			Players []struct {
				ID string  `json:"id"`
				X  float64 `json:"x"`
				Y  float64 `json:"y"`
			} `json:"players"`
		}
		if m.Type != "room.snapshot" && m.Type != "room.joined" {
			continue
		}
		if json.Unmarshal(m.Data, &d) != nil {
			continue
		}
		for _, p := range d.Players {
			if p.ID == id {
				x, y, ok = p.X, p.Y, true
			}
		}
	}
	return x, y, ok
}

// walkToBooth holds keys until the raccoon is past each leg, as in the
// use case steps: left of the fountain, up, then right under the booth.
func (w *smokeWorld) walkToBooth(nickname string) error {
	c, err := w.client(nickname)
	if err != nil {
		return err
	}
	legs := []struct {
		key  string
		done func(x, y float64) bool
	}{
		{"left", func(x, _ float64) bool { return x <= 420 }},
		{"up", func(_, y float64) bool { return y <= 200 }},
		{"right", func(x, _ float64) bool { return x >= 480 }},
	}
	for _, leg := range legs {
		if err := c.send("player.input", map[string]bool{leg.key: true}); err != nil {
			return err
		}
		deadline := time.Now().Add(walkLimit)
		for {
			if x, y, ok := w.position(nickname); ok && leg.done(x, y) {
				break
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("%q did not reach the booth", nickname)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	return c.send("player.input", map[string]bool{})
}

// lastRound waits for a round message and returns the latest one.
func (w *smokeWorld) lastRound(nickname string) (cupsRoundData, error) {
	c, err := w.client(nickname)
	if err != nil {
		return cupsRoundData{}, err
	}
	var r cupsRoundData
	err = c.waitForWithin(walkLimit, "cups.round", func(data json.RawMessage) bool {
		return json.Unmarshal(data, &r) == nil
	})
	return r, err
}

func (w *smokeWorld) bothChoosing(a, b string) error {
	if err := w.joinBoth(a, b); err != nil {
		return err
	}
	// Both walk first: the fast lobby lasts only a second.
	for _, n := range []string{a, b} {
		if err := w.walkToBooth(n); err != nil {
			return err
		}
	}
	for _, n := range []string{a, b} {
		if err := w.clients[n].send("cups.join", map[string]any{}); err != nil {
			return err
		}
	}
	r, err := w.lastRound(a)
	if err != nil {
		return err
	}
	// The round message arrives as the round starts; wait out the shuffle.
	shuffle := time.Duration(r.ShowMs+int64(len(r.Swaps))*r.SwapMs) * time.Millisecond
	time.Sleep(shuffle + 100*time.Millisecond)
	return nil
}

func (w *smokeWorld) pick(nickname string, slot func(cupsRoundData) int) error {
	r, err := w.lastRound(nickname)
	if err != nil {
		return err
	}
	return w.clients[nickname].send("cups.pick", map[string]int{"round": r.Round, "slot": slot(r)})
}

func (w *smokeWorld) picksNut(n string) error {
	return w.pick(n, func(r cupsRoundData) int { return r.nut() })
}

func (w *smokeWorld) picksEmpty(n string) error {
	return w.pick(n, func(r cupsRoundData) int { return (r.nut() + 1) % cups.Slots })
}

func (w *smokeWorld) seesResult(nickname string, hit bool) error {
	c, err := w.client(nickname)
	if err != nil {
		return err
	}
	id := w.ids[nickname]
	return c.waitFor("cups.reveal", func(data json.RawMessage) bool {
		var d cupsRevealData
		if json.Unmarshal(data, &d) != nil {
			return false
		}
		for _, p := range d.Picks {
			if p.ID == id {
				return p.Hit == hit
			}
		}
		return false
	})
}

func (w *smokeWorld) everyoneSeesOnlyHit(nickname string) error {
	id := w.ids[nickname]
	for observer, c := range w.clients {
		err := c.waitFor("cups.reveal", func(data json.RawMessage) bool {
			var d cupsRevealData
			if json.Unmarshal(data, &d) != nil || len(d.Picks) != len(w.clients) {
				return false
			}
			for _, p := range d.Picks {
				if p.Hit != (p.ID == id) {
					return false
				}
			}
			return true
		})
		if err != nil {
			return fmt.Errorf("%q: %w", observer, err)
		}
	}
	return nil
}
