package features

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// heartbeatEvery simulates the transport's ping/pong cadence.
const heartbeatEvery = time.Second

// Steps for conexao.feature.
func (w *world) registerConnectionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a conexão de "([^"]*)" cai$`, w.connectionDrops)
	sc.Step(`^em até (\d+) segundos "([^"]*)" deixa de ver "([^"]*)"$`, w.stopsSeeingWithin)
}

// connectionDrops stops heartbeats for nickname. The socket is not closed
// cleanly, so only the presence timeout can remove them.
func (w *world) connectionDrops(nickname string) error {
	if _, err := w.idOf(nickname); err != nil {
		return err
	}
	w.dropped[nickname] = true
	return nil
}

// stopsSeeingWithin advances time second by second, sending heartbeats for
// every connected player except dropped ones, until the room announces that
// gone left. The observer must still be in the room at the end.
func (w *world) stopsSeeingWithin(seconds, observer, gone string) error {
	limit, err := strconv.Atoi(seconds)
	if err != nil {
		return err
	}
	goneID, err := w.idOf(gone)
	if err != nil {
		return err
	}

	start := w.clock.Now()
	for w.clock.Now().Sub(start) < time.Duration(limit)*time.Second {
		w.clock.Advance(heartbeatEvery)
		for nickname, id := range w.ids {
			if w.dropped[nickname] {
				continue
			}
			if err := w.keepAlive.Execute(id); err != nil {
				return fmt.Errorf("heartbeat for %q: %w", nickname, err)
			}
		}
		if err := w.advance(1); err != nil {
			return err
		}
		if w.announcedLeft(goneID) {
			return w.appearsInRoom(observer)
		}
	}
	return fmt.Errorf("%q still visible after %s seconds", gone, seconds)
}

func (w *world) announcedLeft(id room.PlayerID) bool {
	for _, ev := range w.roomEvents() {
		if left, ok := ev.(usecase.PlayerLeft); ok && left.ID == id {
			return true
		}
	}
	return false
}
