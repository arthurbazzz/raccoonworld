// Package room holds the domain rules of a Raccoons World room: who is in it,
// where they are and what they may say. It has no framework or I/O imports.
package room

import (
	"errors"
	"regexp"
)

// ErrNicknameInvalid is returned when a nickname breaks RN1.
var ErrNicknameInvalid = errors.New("nickname invalid")

// nicknamePattern implements RN1: 3 to 16 ASCII letters, digits or underscores.
var nicknamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)

// Nickname is a validated player name. The zero value is not a valid nickname;
// build one with NewNickname.
type Nickname struct {
	value string
}

// NewNickname validates raw and returns it as a Nickname.
func NewNickname(raw string) (Nickname, error) {
	if !nicknamePattern.MatchString(raw) {
		return Nickname{}, ErrNicknameInvalid
	}
	return Nickname{value: raw}, nil
}

// String returns the nickname exactly as the player typed it.
func (n Nickname) String() string {
	return n.value
}
