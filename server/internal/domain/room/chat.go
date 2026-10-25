package room

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// MaxMessageLength is the chat limit in characters, not bytes (RN4).
	MaxMessageLength = 120
	// MessageInterval is the minimum time between two messages of the
	// same player (RN4).
	MessageInterval = time.Second
)

var (
	// ErrMessageTooLong means the text exceeds MaxMessageLength characters.
	ErrMessageTooLong = errors.New("message too long")
	// ErrMessageEmpty means the text is empty or whitespace only.
	ErrMessageEmpty = errors.New("message empty")
	// ErrRateLimited means the player sent a message less than
	// MessageInterval ago.
	ErrRateLimited = errors.New("message rate limited")
)

// ChatMessage is an accepted chat line, ready to be shown above the sender.
type ChatMessage struct {
	From   PlayerID
	Text   string
	SentAt time.Time
}

// Say validates a chat message from a player. now is passed in by the
// caller so the domain never reads the clock. Rejected messages do not
// count towards the rate limit.
func (r *Room) Say(id PlayerID, text string, now time.Time) (ChatMessage, error) {
	i := r.index(id)
	if i < 0 {
		return ChatMessage{}, ErrPlayerNotFound
	}
	p := r.players[i]

	if strings.TrimSpace(text) == "" {
		return ChatMessage{}, ErrMessageEmpty
	}
	if utf8.RuneCountInString(text) > MaxMessageLength {
		return ChatMessage{}, ErrMessageTooLong
	}
	if !p.lastMessageAt.IsZero() && now.Sub(p.lastMessageAt) < MessageInterval {
		return ChatMessage{}, ErrRateLimited
	}

	p.lastMessageAt = now
	return ChatMessage{From: id, Text: text, SentAt: now}, nil
}
