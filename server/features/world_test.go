package features

import (
	"context"
	"fmt"
	"time"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
	"raccoons-world/server/maps"
)

// roomIDs maps the room names used in the .feature files to room IDs.
var roomIDs = map[string]string{"Praça": "praca"}

// openRoom starts a room loop driven by manual ticks.
func (w *world) openRoom(name string) error {
	id, ok := roomIDs[name]
	if !ok {
		return fmt.Errorf("unknown room %q", name)
	}
	tiles, err := room.ParseTileMap(maps.Praca)
	if err != nil {
		return err
	}
	w.runner = usecase.NewRoomRunner(room.New(id, tiles), w.broadcaster, w.clock)
	w.ticks = make(chan time.Time)
	w.done = make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	w.stop = cancel
	go func() {
		defer close(w.done)
		w.runner.Run(ctx, w.ticks)
	}()

	w.join = usecase.NewJoinRoom(w.runner, w.broadcaster, w.clock)
	w.keepAlive = usecase.NewKeepAlive(w.runner, w.clock)
	w.setInput = usecase.NewSetInput(w.runner)
	w.sendChat = usecase.NewSendChat(w.runner, w.broadcaster, w.clock)
	return nil
}

// advance runs n simulation ticks and waits until the loop processed them.
func (w *world) advance(n int) error {
	for range n {
		w.ticks <- time.Time{}
	}
	return w.runner.Exec(func(*room.Room) {}) // barrier
}

func (w *world) position(nickname string) (room.Vec, error) {
	id, err := w.idOf(nickname)
	if err != nil {
		return room.Vec{}, err
	}
	var (
		p  room.PlayerView
		ok bool
	)
	if err := w.runner.Exec(func(r *room.Room) { p, ok = r.Player(id) }); err != nil {
		return room.Vec{}, err
	}
	if !ok {
		return room.Vec{}, fmt.Errorf("%q is not in the room", nickname)
	}
	return p.Position, nil
}

// ensureJoined joins nickname unless they are already in the room.
func (w *world) ensureJoined(nickname string) error {
	if _, ok := w.ids[nickname]; ok {
		return nil
	}
	return w.mustJoin(nickname)
}

func (w *world) closeRoom() {
	if w.stop == nil {
		return
	}
	w.stop()
	<-w.done
}

// attemptJoin tries to join with a fresh visitor ID and records the outcome.
// A rejection is not a step failure; later steps assert on w.lastErr.
func (w *world) attemptJoin(nickname string) error {
	w.nextVisitor++
	id := room.PlayerID(fmt.Sprintf("visitor-%d", w.nextVisitor))
	w.lastAttempt = id
	w.lastErr = w.join.Execute(usecase.JoinInput{PlayerID: id, Nickname: nickname})
	if w.lastErr == nil {
		w.ids[nickname] = id
	}
	return nil
}

// mustJoin is attemptJoin for steps that expect success.
func (w *world) mustJoin(nickname string) error {
	_ = w.attemptJoin(nickname)
	if w.lastErr != nil {
		return fmt.Errorf("%q could not join: %w", nickname, w.lastErr)
	}
	return nil
}

func (w *world) idOf(nickname string) (room.PlayerID, error) {
	id, ok := w.ids[nickname]
	if !ok {
		return "", fmt.Errorf("%q never joined", nickname)
	}
	return id, nil
}

// players returns the current room state, read on the room goroutine.
func (w *world) players() ([]room.PlayerView, error) {
	var ps []room.PlayerView
	err := w.runner.Exec(func(r *room.Room) { ps = r.Players() })
	return ps, err
}

func (w *world) inRoom(nickname string) (bool, error) {
	ps, err := w.players()
	if err != nil {
		return false, err
	}
	for _, p := range ps {
		if p.Nickname == nickname {
			return true, nil
		}
	}
	return false, nil
}

// eventsTo returns events sent directly to a player.
func (w *world) eventsTo(id room.PlayerID) []usecase.Event {
	var out []usecase.Event
	for _, e := range w.broadcaster.Events() {
		if e.PlayerID == id {
			out = append(out, e.Event)
		}
	}
	return out
}

// roomEvents returns events broadcast to the whole room.
func (w *world) roomEvents() []usecase.Event {
	var out []usecase.Event
	for _, e := range w.broadcaster.Events() {
		if e.RoomID != "" {
			out = append(out, e.Event)
		}
	}
	return out
}
