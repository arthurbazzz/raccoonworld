package room

import "time"

// PresenceTimeout is how long a player may stay silent before being removed
// from the room (RN5). The transport keeps players alive with heartbeats.
const PresenceTimeout = 10 * time.Second

// Touch records a sign of life from a player's connection.
func (r *Room) Touch(id PlayerID, now time.Time) error {
	i := r.index(id)
	if i < 0 {
		return ErrPlayerNotFound
	}
	r.players[i].lastSeenAt = now
	return nil
}

// Expire removes every player silent for PresenceTimeout or longer and
// returns them, in join order.
func (r *Room) Expire(now time.Time) []PlayerView {
	var gone []PlayerView
	kept := r.players[:0]
	for _, p := range r.players {
		if now.Sub(p.lastSeenAt) >= PresenceTimeout {
			gone = append(gone, p.view())
			continue
		}
		kept = append(kept, p)
	}
	clear(r.players[len(kept):]) // drop references to removed players
	r.players = kept
	return gone
}
