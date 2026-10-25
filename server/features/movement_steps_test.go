package features

import (
	"fmt"
	"math"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// direction matches the Portuguese direction words used in movimento.feature.
const direction = `(cima|baixo|a esquerda|a direita)`

// holdTicks is how long "segura a tecla" lasts: 4 ticks = 32px, short
// enough to stay clear of the Praça scenery from the spawn point.
const holdTicks = 4

// untilBlockedTicks bounds walking to a wall (the Praça is 960px wide).
const untilBlockedTicks = 200

// directions maps feature words to the expected sign of (dx, dy).
var directions = map[string][2]int{
	"cima":       {0, -1},
	"baixo":      {0, 1},
	"a esquerda": {-1, 0},
	"a direita":  {1, 0},
}

// Steps for movimento.feature.
func (w *world) registerMovementSteps(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]*)" segura a tecla "([^"]*)"$`, w.holdKey)
	sc.Step(`^"([^"]*)" segura as teclas "([^"]*)" e "([^"]*)"$`, w.holdTwoKeys)
	sc.Step(`^que "([^"]*)" está segurando a tecla "([^"]*)"$`, w.isHoldingKey)
	sc.Step(`^"([^"]*)" solta a tecla "([^"]*)"$`, w.releaseKey)
	sc.Step(`^"([^"]*)" se move para `+direction+`$`, w.movedTowards)
	sc.Step(`^"([^"]*)" vê "([^"]*)" se mover para `+direction+`$`, w.seesMoveTowards)
	sc.Step(`^"([^"]*)" se move para cima e para a direita$`, w.movedUpRight)
	sc.Step(`^a velocidade de "([^"]*)" é a mesma do movimento em linha reta$`, w.straightLineSpeed)
	sc.Step(`^"([^"]*)" para de se mover$`, w.stoppedMoving)
	sc.Step(`^que "([^"]*)" está encostado em um obstáculo à sua direita$`, w.againstObstacleRight)
	sc.Step(`^a posição de "([^"]*)" não atravessa o obstáculo$`, w.didNotPassObstacle)
	sc.Step(`^que "([^"]*)" está na borda esquerda da Praça$`, w.atLeftEdge)
	sc.Step(`^"([^"]*)" continua dentro da Praça$`, w.stillInsideRoom)
}

func withKey(in room.Input, key string, down bool) (room.Input, error) {
	switch key {
	case "W":
		in.Up = down
	case "A":
		in.Left = down
	case "S":
		in.Down = down
	case "D":
		in.Right = down
	default:
		return in, fmt.Errorf("unknown key %q", key)
	}
	return in, nil
}

// changeKeys applies new held keys, marks the current position and runs
// holdTicks so the change has a visible effect.
func (w *world) changeKeys(nickname string, in room.Input) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	if w.mark[nickname], err = w.position(nickname); err != nil {
		return err
	}
	w.held[nickname] = in
	if err := w.setInput.Execute(usecase.SetInputInput{
		PlayerID: id, Up: in.Up, Down: in.Down, Left: in.Left, Right: in.Right,
	}); err != nil {
		return err
	}
	return w.advance(holdTicks)
}

func (w *world) holdKey(nickname, key string) error {
	in, err := withKey(w.held[nickname], key, true)
	if err != nil {
		return err
	}
	return w.changeKeys(nickname, in)
}

func (w *world) holdTwoKeys(nickname, k1, k2 string) error {
	in, err := withKey(w.held[nickname], k1, true)
	if err != nil {
		return err
	}
	if in, err = withKey(in, k2, true); err != nil {
		return err
	}
	return w.changeKeys(nickname, in)
}

func (w *world) isHoldingKey(nickname, key string) error {
	if err := w.ensureJoined(nickname); err != nil {
		return err
	}
	return w.holdKey(nickname, key)
}

func (w *world) releaseKey(nickname, key string) error {
	in, err := withKey(w.held[nickname], key, false)
	if err != nil {
		return err
	}
	return w.changeKeys(nickname, in)
}

// displacement is how far nickname moved since keys last changed.
func (w *world) displacement(nickname string) (room.Vec, error) {
	now, err := w.position(nickname)
	if err != nil {
		return room.Vec{}, err
	}
	start := w.mark[nickname]
	return room.Vec{X: now.X - start.X, Y: now.Y - start.Y}, nil
}

func matchesDirection(d room.Vec, want [2]int) bool {
	return sign(d.X) == want[0] && sign(d.Y) == want[1]
}

func sign(f float64) int {
	switch {
	case f > 0:
		return 1
	case f < 0:
		return -1
	default:
		return 0
	}
}

func (w *world) movedTowards(nickname, dir string) error {
	d, err := w.displacement(nickname)
	if err != nil {
		return err
	}
	if !matchesDirection(d, directions[dir]) {
		return fmt.Errorf("%q moved %+v, want towards %s", nickname, d, dir)
	}
	return nil
}

// seesMoveTowards checks the snapshots broadcast to the room (which the
// observer is in) show the mover going in dir.
func (w *world) seesMoveTowards(observer, mover, dir string) error {
	ok, err := w.inRoom(observer)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("observer %q is not in the room", observer)
	}
	id, err := w.idOf(mover)
	if err != nil {
		return err
	}
	var last *room.Vec
	for _, ev := range w.roomEvents() {
		snap, ok := ev.(usecase.Snapshot)
		if !ok {
			continue
		}
		for _, p := range snap.Players {
			if p.ID == id {
				pos := p.Position
				last = &pos
			}
		}
	}
	if last == nil {
		return fmt.Errorf("no snapshot showed %q moving", mover)
	}
	start := w.mark[mover]
	d := room.Vec{X: last.X - start.X, Y: last.Y - start.Y}
	if !matchesDirection(d, directions[dir]) {
		return fmt.Errorf("snapshots show %q moving %+v, want towards %s", mover, d, dir)
	}
	return nil
}

func (w *world) movedUpRight(nickname string) error {
	d, err := w.displacement(nickname)
	if err != nil {
		return err
	}
	if !matchesDirection(d, [2]int{1, -1}) {
		return fmt.Errorf("%q moved %+v, want up and right", nickname, d)
	}
	return nil
}

func (w *world) straightLineSpeed(nickname string) error {
	d, err := w.displacement(nickname)
	if err != nil {
		return err
	}
	want := room.Speed * usecase.TickInterval.Seconds() * holdTicks
	if got := d.Len(); math.Abs(got-want) > 1e-9 {
		return fmt.Errorf("%q moved %v px, want %v px (same as a straight line)", nickname, got, want)
	}
	return nil
}

func (w *world) stoppedMoving(nickname string) error {
	d, err := w.displacement(nickname)
	if err != nil {
		return err
	}
	if d != (room.Vec{}) {
		return fmt.Errorf("%q kept moving %+v after releasing the keys", nickname, d)
	}
	return nil
}

// walkUntilBlocked holds key until the player stops moving, then releases.
func (w *world) walkUntilBlocked(nickname, key string) error {
	if err := w.ensureJoined(nickname); err != nil {
		return err
	}
	if err := w.holdKey(nickname, key); err != nil {
		return err
	}
	for range untilBlockedTicks {
		before, err := w.position(nickname)
		if err != nil {
			return err
		}
		if err := w.advance(1); err != nil {
			return err
		}
		after, err := w.position(nickname)
		if err != nil {
			return err
		}
		if after == before {
			return w.releaseKey(nickname, key)
		}
	}
	return fmt.Errorf("%q never hit anything holding %q", nickname, key)
}

func (w *world) againstObstacleRight(nickname string) error {
	return w.walkUntilBlocked(nickname, "D")
}

func (w *world) atLeftEdge(nickname string) error {
	return w.walkUntilBlocked(nickname, "A")
}

func (w *world) didNotPassObstacle(nickname string) error {
	return w.stoppedMoving(nickname)
}

func (w *world) stillInsideRoom(nickname string) error {
	if err := w.stoppedMoving(nickname); err != nil {
		return err
	}
	pos, err := w.position(nickname)
	if err != nil {
		return err
	}
	// The left border wall is one 32px tile wide.
	if left := pos.X - room.HitboxWidth/2; left < 32 {
		return fmt.Errorf("%q hitbox left edge at %v, outside the walkable area", nickname, left)
	}
	return nil
}
