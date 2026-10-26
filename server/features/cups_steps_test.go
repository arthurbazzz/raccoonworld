package features

import (
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/cups"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// cupsWorld is the per-scenario state of tres_copos.feature.
type cupsWorld struct {
	round   cups.RoundStarted // last round seen
	roundAt time.Time         // when that round was seen to start
	rounds  int               // CupsRound events seen so far
	tracked int               // where a player who watched the swaps says the nut is
	picks   map[string]int    // nickname -> slot picked in this round
}

// cycleShuffler is deterministic but not trivial: the nut starts in slot 1
// and the swaps cycle through every pair, so the nut moves around.
type cycleShuffler struct{}

func (cycleShuffler) Start() int { return 1 }
func (cycleShuffler) Swaps(n int) []cups.Swap {
	pairs := []cups.Swap{{0, 1}, {1, 2}, {0, 2}}
	out := make([]cups.Swap, n)
	for i := range out {
		out[i] = pairs[i%len(pairs)]
	}
	return out
}

// Steps for tres_copos.feature.
func (w *world) registerCupsSteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a barraca dos três copos está livre$`, w.boothIsFree)
	sc.Step(`^que "([^"]*)" está na Praça longe da barraca$`, w.inRoomAwayFromBooth)
	sc.Step(`^"([^"]*)" anda até perto da barraca$`, w.walkToBooth)
	sc.Step(`^"([^"]*)" vê o aviso "Aperte E para jogar"$`, w.isNearBooth)
	sc.Step(`^"([^"]*)" aperta a tecla "E"$`, w.pressE)
	sc.Step(`^"([^"]*)" não entra em nenhuma partida$`, w.notInAnyMatch)
	sc.Step(`^que "([^"]*)" está perto da barraca$`, w.nearBooth)
	sc.Step(`^começa a espera de (\d+) segundos com "([^"]*)" inscrito$`, w.lobbyOpenedWith)
	sc.Step(`^que "([^"]*)" abriu a espera na barraca$`, w.openedLobby)
	sc.Step(`^"([^"]*)" e "([^"]*)" veem os dois inscritos$`, w.bothSignedUp)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão inscritos na espera$`, w.bothInLobby)
	sc.Step(`^a contagem da espera acaba$`, w.lobbyEnds)
	sc.Step(`^começa a rodada (\d+) de (\d+) para "([^"]*)" e "([^"]*)"$`, w.roundStartsFor)
	sc.Step(`^os dois veem a noz debaixo do mesmo copo$`, w.sameNutForAll)
	sc.Step(`^que "([^"]*)" está jogando os três copos$`, w.playing)
	sc.Step(`^"([^"]*)" vê o aviso "Partida em andamento, você entra na próxima"$`, w.isQueued)
	sc.Step(`^quando a partida de "([^"]*)" acaba, "([^"]*)" está inscrito na próxima espera$`, w.queuedGetsNextLobby)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão na mesma partida$`, w.bothPlaying)
	sc.Step(`^os copos são embaralhados$`, w.cupsShuffled)
	sc.Step(`^"([^"]*)" e "([^"]*)" veem as mesmas trocas$`, w.sameSwapsForAll)
	sc.Step(`^"([^"]*)" acompanha todas as trocas de copos$`, w.followsSwaps)
	sc.Step(`^a noz está no copo que as trocas indicam$`, w.nutWhereSwapsSay)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão numa rodada com o embaralhamento terminado$`, w.bothChoosing)
	sc.Step(`^que "([^"]*)" está numa rodada com o embaralhamento terminado$`, w.choosing)
	sc.Step(`^que "([^"]*)" está numa rodada com os copos ainda trocando$`, w.stillShuffling)
	sc.Step(`^"([^"]*)" escolhe o copo onde está a noz$`, w.picksNut)
	sc.Step(`^"([^"]*)" escolhe um copo sem a noz$`, w.picksEmpty)
	sc.Step(`^"([^"]*)" escolhe um copo$`, w.picksAny)
	sc.Step(`^"([^"]*)" e "([^"]*)" escolhem um copo$`, w.bothPick)
	sc.Step(`^que "([^"]*)" escolheu o copo (\d+)$`, w.picksCup)
	sc.Step(`^"([^"]*)" escolhe o copo (\d+)$`, w.picksCup)
	sc.Step(`^"([^"]*)" vê "Acertou!"$`, w.hit)
	sc.Step(`^"([^"]*)" vê "Errou!"$`, w.missed)
	sc.Step(`^os dois veem que só "([^"]*)" acertou a rodada$`, w.onlyHit)
	sc.Step(`^"([^"]*)" não vê qual copo "([^"]*)" escolheu$`, w.pickHidden)
	sc.Step(`^os copos sobem sem esperar os (\d+) segundos$`, w.revealedEarly)
	sc.Step(`^a escolha é ignorada$`, w.pickIgnored)
	sc.Step(`^a escolha de "([^"]*)" continua sendo o copo (\d+)$`, w.pickStays)
	sc.Step(`^passam (\d+) segundos sem escolha de "([^"]*)"$`, w.secondsPass)
	sc.Step(`^a rodada conta como erro para "([^"]*)"$`, w.countsAsMiss)
	sc.Step(`^"([^"]*)" joga as 5 rodadas$`, w.playsWholeMatch)
	sc.Step(`^cada rodada tem pelo menos tantas trocas quanto a anterior$`, w.swapsNeverDecrease)
	sc.Step(`^as trocas ficam mais rápidas a cada rodada$`, w.swapsGetFaster)
	sc.Step(`^a partida acaba com "([^"]*)" acertando (\d+) rodadas e "([^"]*)" acertando (\d+)$`, w.matchEndsWith)
	sc.Step(`^os dois veem o placar com "([^"]*)" (\d+) de 5 primeiro e "([^"]*)" (\d+) de 5 depois$`, w.scoreboard)
	sc.Step(`^"([^"]*)" aperta a tecla "Esc"$`, w.pressEsc)
	sc.Step(`^"([^"]*)" sai da partida$`, w.leftMatch)
	sc.Step(`^a partida continua para "([^"]*)"$`, w.matchGoesOnFor)
	sc.Step(`^a barraca fica livre$`, w.boothFreed)
	sc.Step(`^em até (\d+) segundos "([^"]*)" sai da partida$`, w.leavesMatchWithin)
	sc.Step(`^que "([^"]*)" está na Praça sem jogar$`, w.ensureJoined)
	sc.Step(`^"([^"]*)" vê os copos se mexendo na barraca$`, w.seesCupsMove)
	sc.Step(`^"([^"]*)" começa uma partida$`, w.openedLobby)
	sc.Step(`^"([^"]*)" vê o sinal de jogando sobre "([^"]*)"$`, w.seesPlayingSign)
	sc.Step(`^a posição de "([^"]*)" não muda$`, w.didNotMove)
}

// --- time and events ---

// passTime advances the clock tick by tick, with heartbeats for everyone
// whose connection did not drop, and notes every round that starts.
func (w *world) passTime(d time.Duration) error {
	end := w.clock.Now().Add(d)
	for w.clock.Now().Before(end) {
		w.clock.Advance(usecase.TickInterval)
		for nickname, id := range w.ids {
			if !w.dropped[nickname] {
				_ = w.keepAlive.Execute(id) // may have left; fine
			}
		}
		if err := w.advance(1); err != nil {
			return err
		}
		w.noteRounds()
	}
	return nil
}

func (w *world) noteRounds() {
	all := cupsEvents[usecase.CupsRound](w)
	if len(all) > w.cups.rounds {
		w.cups.rounds = len(all)
		w.cups.round = all[len(all)-1].Round
		w.cups.roundAt = w.clock.Now()
		w.cups.picks = map[string]int{}
	}
}

// cupsEvents returns the room-wide events of type T, oldest first.
func cupsEvents[T usecase.Event](w *world) []T {
	var out []T
	for _, ev := range w.roomEvents() {
		if v, ok := ev.(T); ok {
			out = append(out, v)
		}
	}
	return out
}

func lastEvent[T usecase.Event](w *world) (T, error) {
	all := cupsEvents[T](w)
	if len(all) == 0 {
		var zero T
		return zero, fmt.Errorf("no %T event", zero)
	}
	return all[len(all)-1], nil
}

func (w *world) status() (cups.StatusChanged, error) {
	s, err := lastEvent[usecase.CupsStatus](w)
	return s.Status, err
}

func (w *world) timing() cups.Timing { return cups.DefaultTiming() }

// --- booth ---

func (w *world) boothIsFree() error {
	if _, err := w.status(); err == nil {
		return fmt.Errorf("the booth has activity before the scenario")
	}
	return nil
}

func (w *world) isNearBooth(nickname string) error {
	pos, err := w.position(nickname)
	if err != nil {
		return err
	}
	var near bool
	if err := w.runner.Exec(func(r *room.Room) { near = r.Map().NearBooth(pos) }); err != nil {
		return err
	}
	if !near {
		return fmt.Errorf("%q at %v is not near the booth", nickname, pos)
	}
	return nil
}

func (w *world) inRoomAwayFromBooth(nickname string) error {
	if err := w.ensureJoined(nickname); err != nil {
		return err
	}
	if w.isNearBooth(nickname) == nil {
		return fmt.Errorf("%q starts near the booth", nickname)
	}
	return nil
}

// walkToBooth walks from the spawn point around the fountain to just below
// the booth: left, up, then right.
func (w *world) walkToBooth(nickname string) error {
	legs := []struct {
		key  string
		done func(room.Vec) bool
	}{
		{"A", func(p room.Vec) bool { return p.X <= 420 }},
		{"W", func(p room.Vec) bool { return p.Y <= 200 }},
		{"D", func(p room.Vec) bool { return p.X >= 480 }},
	}
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	for _, leg := range legs {
		in, err := withKey(room.Input{}, leg.key, true)
		if err != nil {
			return err
		}
		if err := w.setInput.Execute(usecase.SetInputInput{PlayerID: id, Up: in.Up, Down: in.Down, Left: in.Left, Right: in.Right}); err != nil {
			return err
		}
		for range 200 {
			pos, err := w.position(nickname)
			if err != nil {
				return err
			}
			if leg.done(pos) {
				break
			}
			if err := w.advance(1); err != nil {
				return err
			}
		}
	}
	if err := w.setInput.Execute(usecase.SetInputInput{PlayerID: id}); err != nil {
		return err
	}
	w.held[nickname] = room.Input{}
	return w.isNearBooth(nickname)
}

func (w *world) nearBooth(nickname string) error {
	if err := w.ensureJoined(nickname); err != nil {
		return err
	}
	return w.walkToBooth(nickname)
}

func (w *world) pressE(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	return w.joinCups.Execute(id)
}

func (w *world) notInAnyMatch(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	if s, err := w.status(); err == nil && (slices.Contains(s.Players, id) || slices.Contains(s.Queued, id)) {
		return fmt.Errorf("%q is at the booth: %+v", nickname, s)
	}
	for _, ev := range w.eventsTo(id) {
		if r, ok := ev.(usecase.CupsRejected); ok && r.Reason == usecase.CupsTooFar {
			return nil
		}
	}
	return fmt.Errorf("%q was not told they are too far", nickname)
}

// --- lobby ---

func (w *world) openedLobby(nickname string) error {
	if err := w.nearBooth(nickname); err != nil {
		return err
	}
	return w.pressE(nickname)
}

func (w *world) lobbyOpenedWith(seconds int, nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	s, err := w.status()
	if err != nil {
		return err
	}
	if s.Phase != cups.Lobby || !slices.Equal(s.Players, []room.PlayerID{id}) || s.LobbyLeft != time.Duration(seconds)*time.Second {
		return fmt.Errorf("status = %+v, want a %d s lobby with %q", s, seconds, nickname)
	}
	return nil
}

func (w *world) hasPlayers(s cups.StatusChanged, nicknames ...string) error {
	for _, n := range nicknames {
		id, err := w.idOf(n)
		if err != nil {
			return err
		}
		if !slices.Contains(s.Players, id) {
			return fmt.Errorf("%q not among the players: %+v", n, s)
		}
	}
	return nil
}

func (w *world) bothSignedUp(a, b string) error {
	s, err := w.status()
	if err != nil {
		return err
	}
	return w.hasPlayers(s, a, b)
}

func (w *world) bothInLobby(a, b string) error {
	if err := w.openedLobby(a); err != nil {
		return err
	}
	return w.openedLobby(b)
}

func (w *world) lobbyEnds() error { return w.passTime(w.timing().Lobby) }

func (w *world) roundStartsFor(round, total int, a, b string) error {
	r := w.cups.round
	if w.cups.rounds == 0 || r.Round != round || r.Total != total {
		return fmt.Errorf("round = %+v, want %d of %d", r, round, total)
	}
	s, err := w.status()
	if err != nil {
		return err
	}
	if s.Phase != cups.Show {
		return fmt.Errorf("phase = %v, want the nut shown", s.Phase)
	}
	return w.hasPlayers(s, a, b)
}

// sameNutForAll: one round message goes to the whole room, so everyone
// sees the nut under the same cup.
func (w *world) sameNutForAll() error {
	if n := len(cupsEvents[usecase.CupsRound](w)); n != 1 {
		return fmt.Errorf("%d round messages, want 1 for everyone", n)
	}
	return nil
}

// --- playing ---

func (w *world) playing(nickname string) error {
	if err := w.openedLobby(nickname); err != nil {
		return err
	}
	return w.lobbyEnds()
}

func (w *world) bothPlaying(a, b string) error {
	if err := w.bothInLobby(a, b); err != nil {
		return err
	}
	return w.lobbyEnds()
}

func (w *world) isQueued(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	s, err := w.status()
	if err != nil {
		return err
	}
	if !slices.Contains(s.Queued, id) {
		return fmt.Errorf("%q is not queued: %+v", nickname, s)
	}
	return nil
}

func (w *world) queuedGetsNextLobby(_, queued string) error {
	for range 120 { // up to 2 minutes: the match runs out on timeouts
		if err := w.passTime(time.Second); err != nil {
			return err
		}
		if s, _ := w.status(); s.Phase == cups.Lobby {
			return w.hasPlayers(s, queued)
		}
	}
	return fmt.Errorf("the next lobby never opened")
}

// shuffleEnd is when the current round's cups stop moving.
func (w *world) shuffleEnd() time.Time {
	r := w.cups.round
	return w.cups.roundAt.Add(r.Show + time.Duration(len(r.Swaps))*r.SwapTime)
}

// toChoose runs time until the current round's shuffle is over.
func (w *world) toChoose() error {
	if d := w.shuffleEnd().Sub(w.clock.Now()); d > 0 {
		return w.passTime(d)
	}
	return nil
}

func (w *world) cupsShuffled() error { return w.toChoose() }

// sameSwapsForAll: like the nut, the swaps go out once, to the whole room.
func (w *world) sameSwapsForAll(_, _ string) error {
	if len(w.cups.round.Swaps) == 0 {
		return fmt.Errorf("no swaps in round %+v", w.cups.round)
	}
	return w.sameNutForAll()
}

func (w *world) followsSwaps(string) error {
	w.cups.tracked = cups.FinalSlot(w.cups.round.Nut, w.cups.round.Swaps)
	return w.toChoose()
}

func (w *world) nutWhereSwapsSay() error {
	for n := range w.ids {
		if err := w.picksSlot(n, w.cups.tracked); err != nil {
			return err
		}
	}
	rv, err := lastEvent[usecase.CupsRevealed](w)
	if err != nil {
		return err
	}
	if rv.Reveal.Nut != w.cups.tracked {
		return fmt.Errorf("nut under %d, the swaps say %d", rv.Reveal.Nut, w.cups.tracked)
	}
	return nil
}

func (w *world) choosing(nickname string) error {
	if err := w.playing(nickname); err != nil {
		return err
	}
	return w.toChoose()
}

func (w *world) bothChoosing(a, b string) error {
	if err := w.bothPlaying(a, b); err != nil {
		return err
	}
	return w.toChoose()
}

func (w *world) stillShuffling(nickname string) error {
	if err := w.playing(nickname); err != nil {
		return err
	}
	return w.passTime(w.cups.round.Show + usecase.TickInterval)
}

// --- picks ---

func (w *world) nut() int { return cups.FinalSlot(w.cups.round.Nut, w.cups.round.Swaps) }

func (w *world) picksSlot(nickname string, slot int) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	if err := w.pickCup.Execute(usecase.PickInput{PlayerID: id, Round: w.cups.round.Round, Slot: slot}); err != nil {
		return err
	}
	if _, ok := w.cups.picks[nickname]; !ok {
		w.cups.picks[nickname] = slot
	}
	return nil
}

func (w *world) picksNut(nickname string) error { return w.picksSlot(nickname, w.nut()) }
func (w *world) picksEmpty(nickname string) error {
	return w.picksSlot(nickname, (w.nut()+1)%cups.Slots)
}
func (w *world) picksAny(nickname string) error { return w.picksSlot(nickname, 0) }

func (w *world) picksCup(nickname string, cup int) error { return w.picksSlot(nickname, cup-1) }

func (w *world) bothPick(a, b string) error {
	if err := w.picksAny(a); err != nil {
		return err
	}
	return w.picksAny(b)
}

// result returns nickname's line in the last reveal.
func (w *world) result(nickname string) (cups.PickResult, error) {
	id, err := w.idOf(nickname)
	if err != nil {
		return cups.PickResult{}, err
	}
	rv, err := lastEvent[usecase.CupsRevealed](w)
	if err != nil {
		return cups.PickResult{}, err
	}
	for _, p := range rv.Reveal.Picks {
		if p.ID == id {
			return p, nil
		}
	}
	return cups.PickResult{}, fmt.Errorf("%q not in the reveal %+v", nickname, rv)
}

func (w *world) hit(nickname string) error {
	p, err := w.result(nickname)
	if err == nil && !p.Hit {
		err = fmt.Errorf("%q missed: %+v", nickname, p)
	}
	return err
}

func (w *world) missed(nickname string) error {
	p, err := w.result(nickname)
	if err == nil && p.Hit {
		err = fmt.Errorf("%q hit: %+v", nickname, p)
	}
	return err
}

func (w *world) onlyHit(nickname string) error {
	rv, err := lastEvent[usecase.CupsRevealed](w)
	if err != nil {
		return err
	}
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	for _, p := range rv.Reveal.Picks {
		if p.Hit != (p.ID == id) {
			return fmt.Errorf("reveal %+v, want only %q to hit", rv.Reveal, nickname)
		}
	}
	return nil
}

// pickHidden: the room only hears that someone picked, and the cups stay
// down until everyone picks.
func (w *world) pickHidden(_, picker string) error {
	id, err := w.idOf(picker)
	if err != nil {
		return err
	}
	if _, err := lastEvent[usecase.CupsRevealed](w); err == nil {
		return fmt.Errorf("cups went up after one pick")
	}
	for _, p := range cupsEvents[usecase.CupsPicked](w) {
		if p.ID == id {
			return nil // CupsPicked has no cup in it
		}
	}
	return fmt.Errorf("the room was not told %q picked", picker)
}

func (w *world) revealedEarly(seconds int) error {
	if _, err := lastEvent[usecase.CupsRevealed](w); err != nil {
		return err
	}
	if waited := w.clock.Now().Sub(w.shuffleEnd()); waited >= time.Duration(seconds)*time.Second {
		return fmt.Errorf("revealed after %v", waited)
	}
	return nil
}

func (w *world) pickIgnored() error {
	if n := len(cupsEvents[usecase.CupsPicked](w)); n != 0 {
		return fmt.Errorf("%d picks accepted while shuffling", n)
	}
	if err := w.toChoose(); err != nil {
		return err
	}
	if err := w.passTime(w.timing().Pick); err != nil {
		return err
	}
	rv, err := lastEvent[usecase.CupsRevealed](w)
	if err != nil {
		return err
	}
	for _, p := range rv.Reveal.Picks {
		if p.Picked {
			return fmt.Errorf("the early pick counted: %+v", p)
		}
	}
	return nil
}

func (w *world) pickStays(nickname string, cup int) error {
	if err := w.passTime(w.timing().Pick); err != nil { // let the cups go up
		return err
	}
	p, err := w.result(nickname)
	if err != nil {
		return err
	}
	if !p.Picked || p.Slot != cup-1 {
		return fmt.Errorf("%q's pick is %+v, want cup %d", nickname, p, cup)
	}
	return nil
}

func (w *world) secondsPass(seconds int, _ string) error {
	return w.passTime(time.Duration(seconds) * time.Second)
}

func (w *world) countsAsMiss(nickname string) error {
	p, err := w.result(nickname)
	if err == nil && (p.Picked || p.Hit) {
		err = fmt.Errorf("%q: %+v, want a miss without a pick", nickname, p)
	}
	return err
}

// --- whole match ---

// playRounds plays every round left; pick says what each player picks in
// a round (1-based), or -1 to let the time run out.
func (w *world) playRounds(pick func(nickname string, round int) int) error {
	for {
		if err := w.toChoose(); err != nil {
			return err
		}
		round := w.cups.round.Round
		for n := range w.ids {
			if slot := pick(n, round); slot >= 0 {
				if err := w.picksSlot(n, slot); err != nil {
					return err
				}
			}
		}
		// Through the reveal, into the next round or the results.
		before := w.cups.rounds
		for w.cups.rounds == before {
			if _, err := lastEvent[usecase.CupsResults](w); err == nil {
				return nil
			}
			if err := w.passTime(time.Second); err != nil {
				return err
			}
		}
	}
}

func (w *world) playsWholeMatch(string) error {
	return w.playRounds(func(string, int) int { return -1 })
}

func (w *world) swapsNeverDecrease() error {
	rounds := cupsEvents[usecase.CupsRound](w)
	if len(rounds) != 5 {
		return fmt.Errorf("%d rounds, want 5", len(rounds))
	}
	for i := 1; i < len(rounds); i++ {
		if len(rounds[i].Round.Swaps) < len(rounds[i-1].Round.Swaps) {
			return fmt.Errorf("round %d has fewer swaps than round %d", i+1, i)
		}
	}
	return nil
}

func (w *world) swapsGetFaster() error {
	rounds := cupsEvents[usecase.CupsRound](w)
	for i := 1; i < len(rounds); i++ {
		if rounds[i].Round.SwapTime >= rounds[i-1].Round.SwapTime {
			return fmt.Errorf("round %d is not faster than round %d", i+1, i)
		}
	}
	return nil
}

func (w *world) matchEndsWith(a string, hitsA int, b string, hitsB int) error {
	want := map[string]int{a: hitsA, b: hitsB}
	return w.playRounds(func(n string, round int) int {
		if round <= want[n] {
			return w.nut()
		}
		return (w.nut() + 1) % cups.Slots
	})
}

func (w *world) scoreboard(first string, hitsFirst int, second string, hitsSecond int) error {
	res, err := lastEvent[usecase.CupsResults](w)
	if err != nil {
		return err
	}
	idA, _ := w.idOf(first)
	idB, _ := w.idOf(second)
	want := []cups.Score{{ID: idA, Hits: hitsFirst, Place: 1}, {ID: idB, Hits: hitsSecond, Place: 2}}
	if !slices.Equal(res.Results.Scores, want) {
		return fmt.Errorf("scores = %+v, want %+v", res.Results.Scores, want)
	}
	return nil
}

// --- leaving ---

func (w *world) pressEsc(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	return w.leaveCups.Execute(id)
}

func (w *world) leftMatch(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	s, err := w.status()
	if err != nil {
		return err
	}
	if slices.Contains(s.Players, id) {
		return fmt.Errorf("%q still in the match: %+v", nickname, s)
	}
	return nil
}

func (w *world) matchGoesOnFor(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	s, err := w.status()
	if err != nil {
		return err
	}
	if !s.Phase.Playing() || !slices.Equal(s.Players, []room.PlayerID{id}) {
		return fmt.Errorf("status = %+v, want the match going on for %q", s, nickname)
	}
	return nil
}

func (w *world) boothFreed() error {
	s, err := w.status()
	if err != nil {
		return err
	}
	if s.Phase != cups.Idle || len(s.Players) != 0 {
		return fmt.Errorf("status = %+v, want the booth free", s)
	}
	return nil
}

func (w *world) leavesMatchWithin(seconds int, nickname string) error {
	for range seconds {
		if err := w.passTime(time.Second); err != nil {
			return err
		}
		if w.leftMatch(nickname) == nil {
			return nil
		}
	}
	return fmt.Errorf("%q still in the match after %d s", nickname, seconds)
}

// --- seen by others ---

// seesCupsMove: the round, with its swaps, went to the whole room, the
// observer included, while they were not playing.
func (w *world) seesCupsMove(observer string) error {
	id, err := w.idOf(observer)
	if err != nil {
		return err
	}
	if s, _ := w.status(); slices.Contains(s.Players, id) {
		return fmt.Errorf("%q is playing", observer)
	}
	return w.sameSwapsForAll("", "")
}

func (w *world) seesPlayingSign(_, player string) error {
	s, err := w.status()
	if err != nil {
		return err
	}
	return w.hasPlayers(s, player)
}

func (w *world) didNotMove(nickname string) error {
	d, err := w.displacement(nickname)
	if err != nil {
		return err
	}
	if d != (room.Vec{}) {
		return fmt.Errorf("%q moved by %v while playing", nickname, d)
	}
	return nil
}
