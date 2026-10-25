package features

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/usecase"
)

// Steps for chat.feature.
func (w *world) registerChatSteps(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]*)" envia a mensagem "([^"]*)"$`, w.say)
	sc.Step(`^"([^"]*)" vê o balão "([^"]*)" sobre "([^"]*)"$`, w.seesBalloon)
	sc.Step(`^"([^"]*)" acabou de enviar uma mensagem$`, w.justSentMessage)
	sc.Step(`^"([^"]*)" envia outra mensagem em menos de 1 segundo$`, w.sayTooSoon)
	sc.Step(`^a segunda mensagem é descartada$`, w.lastMessageDropped)
	sc.Step(`^"([^"]*)" envia uma mensagem com (\d+) caracteres$`, w.sayWithLength)
	sc.Step(`^a mensagem é descartada$`, w.lastMessageDropped)
}

// say sends a chat line and records the outcome; a rejection is asserted
// by later steps, not here.
func (w *world) say(nickname, text string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	w.lastChatText = text
	w.lastChatErr = w.sendChat.Execute(usecase.SendChatInput{PlayerID: id, Text: text})
	return nil
}

func (w *world) justSentMessage(nickname string) error {
	if err := w.say(nickname, "primeira"); err != nil {
		return err
	}
	if w.lastChatErr != nil {
		return fmt.Errorf("first message rejected: %w", w.lastChatErr)
	}
	return nil
}

func (w *world) sayTooSoon(nickname string) error {
	w.clock.Advance(500 * time.Millisecond)
	return w.say(nickname, "segunda")
}

func (w *world) sayWithLength(nickname, length string) error {
	n, err := strconv.Atoi(length)
	if err != nil {
		return err
	}
	return w.say(nickname, strings.Repeat("a", n))
}

// broadcastChats returns chat lines that reached the room.
func (w *world) broadcastChats() []usecase.ChatMessage {
	var out []usecase.ChatMessage
	for _, ev := range w.roomEvents() {
		if m, ok := ev.(usecase.ChatMessage); ok {
			out = append(out, m)
		}
	}
	return out
}

func (w *world) seesBalloon(observer, text, sender string) error {
	ok, err := w.inRoom(observer)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("observer %q is not in the room", observer)
	}
	id, err := w.idOf(sender)
	if err != nil {
		return err
	}
	for _, m := range w.broadcastChats() {
		if m.From == id && m.Text == text {
			return nil
		}
	}
	return fmt.Errorf("room never got %q from %q", text, sender)
}

func (w *world) lastMessageDropped() error {
	if w.lastChatErr == nil {
		return fmt.Errorf("message %q was accepted", w.lastChatText)
	}
	for _, m := range w.broadcastChats() {
		if m.Text == w.lastChatText {
			return fmt.Errorf("dropped message %q still reached the room", w.lastChatText)
		}
	}
	return nil
}
