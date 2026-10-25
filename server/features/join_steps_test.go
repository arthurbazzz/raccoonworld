package features

import (
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// rejectReasons maps the Portuguese reasons in entrar.feature to protocol values.
var rejectReasons = map[string]usecase.RejectReason{
	"apelido em uso":   usecase.RejectNicknameTaken,
	"apelido inválido": usecase.RejectNicknameInvalid,
	"sala cheia":       usecase.RejectRoomFull,
}

// Steps for entrar.feature.
func (w *world) registerJoinSteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a sala "([^"]*)" está aberta$`, w.openRoom)
	sc.Step(`^o visitante entra com o apelido "([^"]*)"$`, w.mustJoin)
	sc.Step(`^"([^"]*)" entra na Praça$`, w.mustJoin)
	sc.Step(`^que "([^"]*)" está na Praça$`, w.mustJoin)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão na Praça$`, w.bothJoin)
	sc.Step(`^outro visitante tenta entrar com o apelido "([^"]*)"$`, w.attemptJoin)
	sc.Step(`^o visitante tenta entrar com o apelido "([^"]*)"$`, w.attemptJoin)
	sc.Step(`^que a Praça está com a capacidade máxima$`, w.fillRoom)
	sc.Step(`^"([^"]*)" aparece na Praça$`, w.appearsInRoom)
	sc.Step(`^"([^"]*)" recebe a lista dos jogadores já presentes$`, w.receivesPlayerList)
	sc.Step(`^a entrada é recusada com o motivo "([^"]*)"$`, w.joinRejectedWith)
	sc.Step(`^"([^"]*)" é avisado de que "([^"]*)" entrou$`, w.isNotifiedOfArrival)
}

func (w *world) bothJoin(a, b string) error {
	if err := w.mustJoin(a); err != nil {
		return err
	}
	return w.mustJoin(b)
}

func (w *world) fillRoom() error {
	for i := range room.MaxPlayers {
		if err := w.mustJoin(fmt.Sprintf("filler%d", i)); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) appearsInRoom(nickname string) error {
	ok, err := w.inRoom(nickname)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%q is not in the room", nickname)
	}
	return nil
}

func (w *world) receivesPlayerList(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	current, err := w.players()
	if err != nil {
		return err
	}
	for _, ev := range w.eventsTo(id) {
		joined, ok := ev.(usecase.Joined)
		if !ok {
			continue
		}
		if joined.You.ID != id {
			return fmt.Errorf("joined.You = %q, want %q", joined.You.ID, id)
		}
		if len(joined.Players) != len(current) {
			return fmt.Errorf("player list has %d players, room has %d", len(joined.Players), len(current))
		}
		return nil
	}
	return fmt.Errorf("%q never received the Joined event", nickname)
}

func (w *world) joinRejectedWith(reason string) error {
	want, ok := rejectReasons[reason]
	if !ok {
		return fmt.Errorf("unknown reason %q", reason)
	}
	var rejected *usecase.JoinRejectedError
	if !errors.As(w.lastErr, &rejected) || rejected.Reason != want {
		return fmt.Errorf("join error = %w, want rejection %q", w.lastErr, want)
	}
	for _, ev := range w.eventsTo(w.lastAttempt) {
		if r, ok := ev.(usecase.JoinRejected); ok && r.Reason == want {
			return nil
		}
	}
	return fmt.Errorf("visitor never received JoinRejected(%q)", want)
}

func (w *world) isNotifiedOfArrival(observer, newcomer string) error {
	ok, err := w.inRoom(observer)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("observer %q is not in the room", observer)
	}
	for _, ev := range w.roomEvents() {
		if pj, ok := ev.(usecase.PlayerJoined); ok && pj.Player.Nickname == newcomer {
			return nil
		}
	}
	return fmt.Errorf("room never got PlayerJoined for %q", newcomer)
}
