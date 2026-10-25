package usecase

import "raccoons-world/server/internal/domain/room"

// SendChatInput is a chat line typed by a player.
type SendChatInput struct {
	PlayerID room.PlayerID
	Text     string
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

// Execute broadcasts the line as a ChatMessage. Rejected lines (too long,
// blank, rate limited) are dropped silently: nobody is notified, and the
// domain error is returned for logging only.
func (uc *SendChat) Execute(in SendChatInput) error {
	now := uc.clock.Now()
	var sayErr error
	err := uc.runner.Exec(func(r *room.Room) {
		var msg room.ChatMessage
		msg, sayErr = r.Say(in.PlayerID, in.Text, now)
		if sayErr != nil {
			return
		}
		uc.broadcaster.ToRoom(r.ID(), ChatMessage{From: msg.From, Text: msg.Text})
	})
	if err != nil {
		return err
	}
	return sayErr
}
