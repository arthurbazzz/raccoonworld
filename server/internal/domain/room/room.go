package room

import (
	"errors"
	"slices"
	"time"
)

const (
	// MaxPlayers is the room capacity (RN2).
	MaxPlayers = 10
	// Speed is how fast a raccoon walks, in pixels per second (RN3).
	Speed = 160.0
)

var (
	// ErrNicknameTaken means another player in the room has the same
	// nickname, ignoring case (RN1).
	ErrNicknameTaken = errors.New("nickname taken")
	// ErrRoomFull means the room already has MaxPlayers players (RN2).
	ErrRoomFull = errors.New("room full")
	// ErrPlayerExists means the player ID is already in the room.
	ErrPlayerExists = errors.New("player already in room")
	// ErrPlayerNotFound means the player ID is not in the room.
	ErrPlayerNotFound = errors.New("player not in room")
)

// Room is the authoritative state of one room (constitution, principle V).
// It is not safe for concurrent use: a single goroutine owns each room.
type Room struct {
	id      string
	tiles   *TileMap
	players []*player // join order
}

// New creates an empty room using tiles for collision and spawning.
func New(id string, tiles *TileMap) *Room {
	return &Room{id: id, tiles: tiles}
}

// ID returns the room identifier.
func (r *Room) ID() string { return r.id }

// Map returns the room's tile map.
func (r *Room) Map() *TileMap { return r.tiles }

// Join adds a player at the spawn point. Capacity is checked before the
// nickname, so a full room always answers ErrRoomFull.
func (r *Room) Join(id PlayerID, nickname Nickname) (PlayerView, error) {
	if len(r.players) >= MaxPlayers {
		return PlayerView{}, ErrRoomFull
	}
	for _, p := range r.players {
		if p.id == id {
			return PlayerView{}, ErrPlayerExists
		}
		if p.nickname.Key() == nickname.Key() {
			return PlayerView{}, ErrNicknameTaken
		}
	}
	p := &player{id: id, nickname: nickname, position: r.tiles.Spawn()}
	r.players = append(r.players, p)
	return p.view(), nil
}

// Leave removes a player and reports whether they were in the room.
func (r *Room) Leave(id PlayerID) bool {
	i := r.index(id)
	if i < 0 {
		return false
	}
	r.players = slices.Delete(r.players, i, i+1)
	return true
}

// SetInput replaces the keys a player is holding.
func (r *Room) SetInput(id PlayerID, in Input) error {
	i := r.index(id)
	if i < 0 {
		return ErrPlayerNotFound
	}
	r.players[i].input = in
	return nil
}

// Player returns a snapshot of one player.
func (r *Room) Player(id PlayerID) (PlayerView, bool) {
	i := r.index(id)
	if i < 0 {
		return PlayerView{}, false
	}
	return r.players[i].view(), true
}

// Players returns snapshots of everyone in join order.
func (r *Room) Players() []PlayerView {
	views := make([]PlayerView, len(r.players))
	for i, p := range r.players {
		views[i] = p.view()
	}
	return views
}

// Step advances movement by dt and returns the players whose position
// changed. It never reads the clock, so it is deterministic in tests.
func (r *Room) Step(dt time.Duration) []PlayerView {
	var moved []PlayerView
	for _, p := range r.players {
		dir := p.input.Direction()
		if dir == (Vec{}) {
			continue
		}
		before := p.position
		p.setHitbox(r.tiles.Move(p.hitbox(), dir.Scale(Speed*dt.Seconds())))
		if p.position != before {
			moved = append(moved, p.view())
		}
	}
	return moved
}

func (r *Room) index(id PlayerID) int {
	return slices.IndexFunc(r.players, func(p *player) bool { return p.id == id })
}
