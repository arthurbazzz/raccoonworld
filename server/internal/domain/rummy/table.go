package rummy

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"raccoons-world/server/internal/domain/room"
)

// Errors returned by Table.Create and Table.Join (sent to the client as the
// reason of a rejection).
var (
	ErrSeated  = errors.New("busy")    // already in a room of this table
	ErrNoRoom  = errors.New("no_room") // no room with that ID
	ErrFull    = errors.New("full")
	ErrPlaying = errors.New("playing")
	ErrNoMode  = errors.New("bad_mode")
)

// Phase is where a room of the table is (RN5-RN7).
type Phase int

// The phases of a room, in order.
const (
	Lobby   Phase = iota // waiting for players
	Playing              // a match is on
	Results              // final scoreboard
)

func (p Phase) String() string {
	return [...]string{"lobby", "playing", "results"}[p]
}

type tableRoom struct {
	id        string
	mode      Mode
	players   []Player // in join order
	phase     Phase
	lobbyEnds time.Time // zero while fewer than MinPlayers wait
	match     *Match
	again     map[room.PlayerID]bool // asked to play again from the scoreboard
	resultsAt time.Time
}

// Table is the pieces table in the Praça: a list of rooms, each with up to
// four players and a match of its own (RN5).
type Table struct {
	timing   Timing
	shuffler Shuffler
	rooms    []*tableRoom
	nextID   int
}

// NewTable returns an empty table.
func NewTable(t Timing, s Shuffler) *Table {
	return &Table{timing: t, shuffler: s}
}

// RoomInfo is a room as the room list shows it (RN5).
type RoomInfo struct {
	ID        string
	Mode      Mode
	Players   []Player
	Phase     Phase
	LobbyEnds time.Time // zero when there is no countdown
}

// Rooms lists the rooms, oldest first.
func (t *Table) Rooms() []RoomInfo {
	out := make([]RoomInfo, len(t.rooms))
	for i, r := range t.rooms {
		players := slices.Clone(r.players)
		if r.match != nil {
			players = r.match.Players()
		}
		out[i] = RoomInfo{ID: r.id, Mode: r.mode, Players: players, Phase: r.phase, LobbyEnds: r.lobbyEnds}
	}
	return out
}

// Playing lists who is in a room of the table, waiting or playing (RN27).
func (t *Table) Playing() []room.PlayerID {
	var out []room.PlayerID
	for _, r := range t.rooms {
		for _, p := range t.roomPlayers(r) {
			out = append(out, p.ID)
		}
	}
	return out
}

// Matches is how many rooms are playing (RN25).
func (t *Table) Matches() int {
	n := 0
	for _, r := range t.rooms {
		if r.phase == Playing {
			n++
		}
	}
	return n
}

func (t *Table) roomPlayers(r *tableRoom) []Player {
	if r.phase == Playing {
		return r.match.Players()
	}
	return r.players
}

func (t *Table) find(id string) *tableRoom {
	for _, r := range t.rooms {
		if r.id == id {
			return r
		}
	}
	return nil
}

// RoomOf returns the ID of the room id is in.
func (t *Table) RoomOf(id room.PlayerID) (string, bool) {
	for _, r := range t.rooms {
		if slices.ContainsFunc(t.roomPlayers(r), func(p Player) bool { return p.ID == id }) {
			return r.id, true
		}
	}
	return "", false
}

// Match returns the match of a room, if it is playing or showing results.
func (t *Table) Match(roomID string) (*Match, bool) {
	r := t.find(roomID)
	if r == nil || r.match == nil {
		return nil, false
	}
	return r.match, true
}

// Create opens a room with p in it, in the given mode (RN5).
func (t *Table) Create(p Player, mode Mode) (string, []Event, error) {
	if mode != Classic && mode != Free {
		return "", nil, ErrNoMode
	}
	if _, ok := t.RoomOf(p.ID); ok {
		return "", nil, ErrSeated
	}
	t.nextID++
	r := &tableRoom{id: fmt.Sprintf("s%d", t.nextID), mode: mode, players: []Player{p}}
	t.rooms = append(t.rooms, r)
	return r.id, []Event{RoomsChanged{}}, nil
}

// Join puts p in a waiting room with a free seat (RN5-RN7). The countdown
// starts with 2 players; with 4 the match starts at once.
func (t *Table) Join(p Player, roomID string, now time.Time) ([]Event, error) {
	if _, ok := t.RoomOf(p.ID); ok {
		return nil, ErrSeated
	}
	r := t.find(roomID)
	switch {
	case r == nil:
		return nil, ErrNoRoom
	case r.phase != Lobby:
		return nil, ErrPlaying
	case len(r.players) >= MaxPlayers:
		return nil, ErrFull
	}
	r.players = append(r.players, p)
	if len(r.players) == MaxPlayers {
		return append([]Event{RoomsChanged{}}, t.start(r, now)...), nil
	}
	if r.lobbyEnds.IsZero() && len(r.players) >= MinPlayers {
		r.lobbyEnds = now.Add(t.timing.Lobby)
	}
	return []Event{RoomsChanged{}}, nil
}

func (t *Table) start(r *tableRoom, now time.Time) []Event {
	m, events := NewMatch(r.mode, r.players, t.shuffler, t.timing, now)
	r.match, r.phase, r.lobbyEnds = m, Playing, time.Time{}
	out := []Event{MatchStarted{Room: r.id, Players: m.Players()}}
	return append(out, t.wrap(r, events, now)...)
}

// PlayAgain keeps id in the room for the next match, from the scoreboard (RN7).
func (t *Table) PlayAgain(id room.PlayerID) {
	for _, r := range t.rooms {
		if r.phase == Results && slices.ContainsFunc(r.players, func(p Player) bool { return p.ID == id }) {
			r.again[id] = true
		}
	}
}

// Leave takes id out of the table: out of the lobby, the match or the
// scoreboard (RN20). An empty room goes away (RN5).
func (t *Table) Leave(id room.PlayerID, now time.Time) []Event {
	roomID, ok := t.RoomOf(id)
	if !ok {
		return nil
	}
	r := t.find(roomID)
	var events []Event
	if r.phase == Playing {
		events = t.wrap(r, r.match.Leave(id, now), now)
	} else {
		r.players = slices.DeleteFunc(r.players, func(p Player) bool { return p.ID == id })
		delete(r.again, id)
		if len(r.players) < MinPlayers {
			r.lobbyEnds = time.Time{}
		}
	}
	events = append(events, Released{IDs: []room.PlayerID{id}})
	t.dropEmpty()
	return append(events, RoomsChanged{})
}

func (t *Table) dropEmpty() {
	t.rooms = slices.DeleteFunc(t.rooms, func(r *tableRoom) bool { return len(t.roomPlayers(r)) == 0 })
}

// SetAbsent marks a dropped player (RN21). Only a match keeps the seat; in a
// lobby or on the scoreboard the player simply leaves.
func (t *Table) SetAbsent(id room.PlayerID, now time.Time) []Event {
	roomID, ok := t.RoomOf(id)
	if !ok {
		return nil
	}
	if r := t.find(roomID); r.phase == Playing {
		r.match.SetAbsent(id, now)
		return []Event{InMatch{Room: r.id}}
	}
	return t.Leave(id, now)
}

// Reclaim gives an absent seat back to a player who returned with the same
// nickname under a new ID (plan, section 3.2).
func (t *Table) Reclaim(nickname string, id room.PlayerID) (string, room.PlayerID, bool) {
	for _, r := range t.rooms {
		if r.phase != Playing {
			continue
		}
		if old, ok := r.match.Reclaim(nickname, id); ok {
			return r.id, old, true
		}
	}
	return "", "", false
}

// Advance runs the countdowns, the matches and the scoreboards.
func (t *Table) Advance(now time.Time) []Event {
	var events []Event
	for _, r := range slices.Clone(t.rooms) {
		switch r.phase {
		case Lobby:
			if !r.lobbyEnds.IsZero() && !now.Before(r.lobbyEnds) {
				events = append(events, t.start(r, now)...)
				events = append(events, RoomsChanged{})
			}
		case Playing:
			events = append(events, t.wrap(r, r.match.Advance(now), now)...)
		case Results:
			if !now.Before(r.resultsAt.Add(t.timing.Results)) {
				events = append(events, t.reopen(r, now)...)
			}
		}
	}
	t.dropEmpty()
	return events
}

// wrap tags match events with the room, and moves a room whose match ended
// to the scoreboard. Players who left the match are released.
func (t *Table) wrap(r *tableRoom, events []Event, now time.Time) []Event {
	out := make([]Event, 0, len(events)+1)
	for _, e := range events {
		out = append(out, InMatch{Room: r.id, Event: e})
		switch e := e.(type) {
		case PlayerRemoved:
			out = append(out, Released{IDs: []room.PlayerID{e.Player.ID}})
		case MatchOver:
			r.phase = Results
			r.players = r.match.Players()
			r.again = map[room.PlayerID]bool{}
			out = append(out, RoomsChanged{})
		}
	}
	if r.phase == Results && r.resultsAt.IsZero() {
		r.resultsAt = now
	}
	return out
}

// reopen ends the scoreboard: who asked to play again waits in the same
// room, everyone else is released (RN7).
func (t *Table) reopen(r *tableRoom, now time.Time) []Event {
	var gone []room.PlayerID
	kept := r.players[:0]
	for _, p := range r.players {
		if r.again[p.ID] {
			kept = append(kept, p)
		} else {
			gone = append(gone, p.ID)
		}
	}
	r.players, r.phase, r.match, r.again, r.resultsAt = kept, Lobby, nil, nil, time.Time{}
	if len(r.players) >= MinPlayers {
		r.lobbyEnds = now.Add(t.timing.Lobby)
	}
	var events []Event
	if len(gone) > 0 {
		events = append(events, Released{IDs: gone})
	}
	return append(events, RoomsChanged{})
}
