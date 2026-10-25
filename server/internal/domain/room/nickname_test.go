package room_test

import (
	"errors"
	"testing"

	"raccoons-world/server/internal/domain/room"
)

func TestNewNickname(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"minimum length", "abc", false},
		{"maximum length", "abcdefghijklmnop", false},
		{"letters digits underscore", "Rocky_42", false},
		{"only digits", "123", false},
		{"too short", "ab", true},
		{"single char", "a", true},
		{"empty", "", true},
		{"too long", "abcdefghijklmnopq", true},
		{"space", "com espaco", true},
		{"symbol", "rocky!", true},
		{"hyphen", "rocky-1", true},
		{"accented letter", "guaxinimé", true},
		{"leading space", " rocky", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := room.NewNickname(tt.input)
			if tt.wantErr {
				if !errors.Is(err, room.ErrNicknameInvalid) {
					t.Fatalf("NewNickname(%q) error = %v, want ErrNicknameInvalid", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewNickname(%q) unexpected error: %v", tt.input, err)
			}
			if got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
		})
	}
}
