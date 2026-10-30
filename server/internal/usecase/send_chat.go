package usecase

import "raccoons-world/server/internal/domain/room"

// SendChatInput is a chat line typed by a player. Table lines go only to
// the player's match at the pieces table (RW-006, RN37).
type SendChatInput struct {
	PlayerID room.PlayerID
	Text     string
	Table    bool
}

// SendChat validates a chat line and broadcasts it to the room (RN4).
type SendChat struct {
	runner      *RoomRunner
	broadcaster Broadcaster
	clock       Clock
}

// NewSendChat builds the use case for one room.
func NewSendChat(runner *RoomRunner, b Broadcaster, c Clock) *SendChat {
	return &SendChat{runner: runner, broadcaster: b, clock: c}
}

// Execute broadcasts the line as a ChatMessage, or a RummyChat for a table
// line. Rejected lines (too long, blank, rate limited, table line from
// someone not in a match) are dropped silently: nobody is notified, and the
// domain error is returned for logging only. Both kinds share the same
// limits (RN39).
func (uc *SendChat) Execute(in SendChatInput) error {
	now := uc.clock.Now()
	var sayErr error
	err := uc.runner.Exec(func(r *room.Room) {
		var mates []room.PlayerID
		if in.Table {
			var ok bool
			if mates, ok = uc.runner.matchMates(in.PlayerID); !ok {
				sayErr = ErrNotInMatch
				return
			}
		}
		var msg room.ChatMessage
		msg, sayErr = r.Say(in.PlayerID, in.Text, now)
		if sayErr != nil {
			return
		}
		if !in.Table {
			uc.broadcaster.ToRoom(r.ID(), ChatMessage{From: msg.From, Text: msg.Text})
			return
		}
		for _, id := range mates {
			uc.broadcaster.ToPlayer(id, RummyChat{From: msg.From, Text: msg.Text})
		}
	})
	if err != nil {
		return err
	}
	return sayErr
}
