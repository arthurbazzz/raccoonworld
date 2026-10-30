package features

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/domain/rummy"
	"raccoons-world/server/internal/usecase"
)

// rummyWorld is the per-scenario state of rummy.feature.
type rummyWorld struct {
	shuffler   riggedShuffler
	tableSeq   string              // "a mesa tem a sequência": the opener of the next setup
	start      rummy.View          // the current player's view at the start of the turn
	startHands map[string][]string // nickname -> faces in hand, before a disconnect
	lastMove   struct {            // the last move sent, which a refusal keeps
		turn  int
		melds [][]rummy.TileID
		hand  []rummy.TileID
	}
}

// riggedShuffler deals the deck it is given, in order; with no deck, or for
// anything but a full deck, it leaves the order as is. Seats then follow the
// join order, so the room creator plays first.
type riggedShuffler struct{ deck []rummy.TileID }

func (s *riggedShuffler) Shuffle(tiles []rummy.TileID) {
	if len(tiles) == rummy.DeckSize && s.deck != nil {
		copy(tiles, s.deck)
	}
}

// rig makes the next deal give these faces to the first seats, padded to 14
// with tiles from the end of the deck (jokers stay in the pool).
func (s *riggedShuffler) rig(hands ...[]string) error {
	used := map[rummy.TileID]bool{}
	take := func(face string) (rummy.TileID, error) {
		for _, id := range rummy.NewDeck() {
			if !used[id] && rummy.TileOf(id).String() == face {
				used[id] = true
				return id, nil
			}
		}
		return 0, fmt.Errorf("no tile %s left to deal", face)
	}
	ids := make([][]rummy.TileID, len(hands))
	for i, hand := range hands {
		for _, face := range hand {
			id, err := take(face)
			if err != nil {
				return err
			}
			ids[i] = append(ids[i], id)
		}
	}
	var deck []rummy.TileID
	next := rummy.TileID(2*rummy.Colors*rummy.MaxValue - 1)
	for _, hand := range ids {
		for len(hand) < rummy.HandSize {
			if !used[next] {
				used[next] = true
				hand = append(hand, next)
			}
			next--
		}
		deck = append(deck, hand...)
	}
	for _, id := range rummy.NewDeck() {
		if !used[id] {
			deck = append(deck, id)
		}
	}
	s.deck = deck
	return nil
}

// Steps for rummy.feature.
func (w *world) registerRummySteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a mesa de peças está livre$`, w.tableIsFree)
	sc.Step(`^que "([^"]*)" está perto da mesa$`, w.nearTable)
	sc.Step(`^"([^"]*)" cria uma sala no modo "([^"]*)"$`, w.createRoom)
	sc.Step(`^que "([^"]*)" criou uma sala no modo "([^"]*)"$`, w.createdRoom)
	sc.Step(`^começa a espera de uma sala no modo "([^"]*)" com "([^"]*)" inscrito$`, w.lobbyWith)
	sc.Step(`^"([^"]*)" abre a lista de salas perto da mesa$`, w.opensList)
	sc.Step(`^"([^"]*)" vê a sala de "([^"]*)" no modo "([^"]*)" com (\d+) vagas$`, w.seesRoomWithSeats)
	sc.Step(`^quando "([^"]*)" entra na sala de "([^"]*)", os dois estão na mesma espera$`, w.joinsSameLobby)
	sc.Step(`^"([^"]*)" entra na sala de "([^"]*)"$`, w.joinRoomOf)
	sc.Step(`^começa a contagem de (\d+) segundos na sala de "([^"]*)"$`, w.countdownIn)
	sc.Step(`^que "([^"]*)", "([^"]*)" e "([^"]*)" estão na espera da mesma sala$`, w.threeInLobby)
	sc.Step(`^a partida da sala de "([^"]*)" começa com (\d+) jogadores$`, w.matchStartsWith)
	sc.Step(`^que "([^"]*)", "([^"]*)", "([^"]*)" e "([^"]*)" estão jogando na mesma sala$`, w.fourPlaying)
	sc.Step(`^que "([^"]*)", "([^"]*)" e "([^"]*)" estão jogando na mesma sala$`, w.threePlaying)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão jogando na mesma sala$`, w.twoPlaying)
	sc.Step(`^"([^"]*)" e "([^"]*)" começam uma partida em outra sala$`, w.twoPlaying)
	sc.Step(`^começa uma partida entre "([^"]*)" e "([^"]*)"$`, w.twoPlaying)
	sc.Step(`^as duas salas estão jogando$`, w.twoMatches)
	sc.Step(`^"([^"]*)" não recebe a mesa da sala de "([^"]*)"$`, w.noStateFrom)
	sc.Step(`^"([^"]*)" vê a sala de "([^"]*)" como "jogando"$`, w.seesPlaying)
	sc.Step(`^"([^"]*)" não consegue entrar na sala de "([^"]*)"$`, w.cannotJoin)
	sc.Step(`^cada um tem (\d+) peças na mão$`, w.everyHandHas)
	sc.Step(`^o monte tem (\d+) peças$`, w.poolHas)
	sc.Step(`^"([^"]*)" sabe quantas peças "([^"]*)" tem, mas não quais$`, w.handIsSecret)

	sc.Step(`^que é a vez de "([^"]*)", que ainda não abriu, no modo "([^"]*)"$`, w.turnNotOpened)
	sc.Step(`^que a mesa tem a sequência "([^"]*)"$`, w.tableHasRun)
	sc.Step(`^que é a vez de "([^"]*)", que já abriu, com "([^"]*)" na mão$`, w.turnOpenedWith)
	sc.Step(`^que é a vez de "([^"]*)", que já abriu, só com "([^"]*)" na mão$`, w.turnOpenedOnly)
	sc.Step(`^"([^"]*)" baixa (?:a sequência|o grupo) "([^"]*)" e termina a vez$`, w.layAndEnd)
	sc.Step(`^que "([^"]*)" baixou o grupo "([^"]*)" sem terminar a vez$`, w.lay)
	sc.Step(`^"([^"]*)" tira o "([^"]*)" da sequência e forma o grupo "([^"]*)"$`, w.takeFromRun)
	sc.Step(`^"([^"]*)" termina a vez$`, w.endTurn)
	sc.Step(`^"([^"]*)" desfaz a vez$`, w.undoTurn)
	sc.Step(`^a jogada de "([^"]*)" é recusada com o motivo "([^"]*)"$`, w.playRefusedWith)
	sc.Step(`^as peças "([^"]*)" ficam destacadas$`, w.tilesBlamed)
	sc.Step(`^ainda é a vez de "([^"]*)"$`, w.stillTurnOf)
	sc.Step(`^"([^"]*)" não compra peça$`, w.didNotDraw)
	sc.Step(`^que a jogada "([^"]*)" de "([^"]*)" foi recusada$`, w.refusedPlay)
	sc.Step(`^"([^"]*)" junta o "([^"]*)" à combinação e termina a vez$`, w.addToLastAndEnd)
	sc.Step(`^o tempo da vez acaba$`, w.turnTimeRunsOut)
	sc.Step(`^"([^"]*)" compra uma peça automaticamente$`, w.timedOutAndDrewTile)
	sc.Step(`^"([^"]*)" arrasta a sequência "([^"]*)" para a posição (\d+), (\d+)$`, w.moveMeldTo)
	sc.Step(`^"([^"]*)" vê a sequência "([^"]*)" na posição (\d+), (\d+)$`, w.seesMeldAt)
	sc.Step(`^"([^"]*)" envia "([^"]*)" no chat da mesa$`, w.tableSay)
	sc.Step(`^"([^"]*)" recebe "([^"]*)" de "([^"]*)" no chat da mesa$`, w.receivesTableChat)
	sc.Step(`^"([^"]*)" não recebe mensagens do chat da mesa$`, w.noTableChat)
	sc.Step(`^a mesa volta ao começo da vez$`, w.tableBackToStart)
	sc.Step(`^as peças de "([^"]*)" ficam na mesa$`, w.playAccepted)
	sc.Step(`^é a vez do próximo jogador$`, w.nextPlayersTurn)
	sc.Step(`^a mesa tem a sequência "([^"]*)" e o grupo "([^"]*)"$`, w.tableHolds)
	sc.Step(`^a mesa e a mão de "([^"]*)" voltam ao começo da vez$`, w.backToStart)
	sc.Step(`^passam (\d+) segundos sem "([^"]*)" terminar a vez$`, w.turnRunsOut)
	sc.Step(`^a vez de "([^"]*)" é desfeita e ele compra uma peça$`, w.timedOutAndDrew)
	sc.Step(`^todos na sala veem que "([^"]*)" ganhou$`, w.everyoneSeesWinner)
	sc.Step(`^veem o placar com os pontos que sobraram na mão de cada um$`, w.scoreboardWithPoints)
	sc.Step(`^"([^"]*)" ganha a partida$`, w.wins)

	sc.Step(`^"([^"]*)" sai da Praça e volta com o mesmo apelido em (\d+) segundos$`, w.dropAndReturn)
	sc.Step(`^"([^"]*)" continua na partida com a mesma mão$`, w.sameHandAfterReturn)
	sc.Step(`^"([^"]*)" sai da Praça e não volta por (\d+) segundos$`, w.dropForGood)
	sc.Step(`^"([^"]*)" sai da partida e as peças dele voltam ao monte$`, w.removedTilesBack)
	sc.Step(`^a partida continua para "([^"]*)" e "([^"]*)"$`, w.matchGoesOnForTwo)
	sc.Step(`^"([^"]*)" sai da mesa$`, w.leaveTable)
}

// --- faces ---

var colorWords = map[string]string{"marrom": "brown", "vermelho": "red", "verde": "green", "azul": "blue"}

// faces turns the scenario notation into tile faces: "azul 3-4-5" is a
// run; "10 vermelho, 10 azul, coringa" a list.
func faces(spec string) ([]string, error) {
	if color, nums, ok := strings.Cut(spec, " "); ok && strings.Contains(nums, "-") {
		en, ok := colorWords[color]
		if !ok {
			return nil, fmt.Errorf("unknown color %q", color)
		}
		var out []string
		for _, n := range strings.Split(nums, "-") {
			out = append(out, en+"-"+n)
		}
		return out, nil
	}
	var out []string
	for _, part := range strings.Split(spec, ", ") {
		if part == "coringa" {
			out = append(out, "joker")
			continue
		}
		n, color, ok := strings.Cut(part, " ")
		en, known := colorWords[color]
		if _, err := strconv.Atoi(n); !ok || !known || err != nil {
			return nil, fmt.Errorf("cannot read tile %q", part)
		}
		out = append(out, en+"-"+n)
	}
	return out, nil
}

func faceOf(id rummy.TileID) string { return rummy.TileOf(id).String() }

func facesOf(ids []rummy.TileID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = faceOf(id)
	}
	return out
}

// pick returns IDs from hand with these faces, in this order.
func pick(hand []rummy.TileID, want []string) ([]rummy.TileID, []rummy.TileID, error) {
	left := slices.Clone(hand)
	var out []rummy.TileID
	for _, f := range want {
		i := slices.IndexFunc(left, func(id rummy.TileID) bool { return faceOf(id) == f })
		if i < 0 {
			return nil, nil, fmt.Errorf("%s is not in %v", f, facesOf(hand))
		}
		out = append(out, left[i])
		left = slices.Delete(left, i, i+1)
	}
	return out, left, nil
}

// --- events ---

// rummyTo returns the events of type T sent straight to a player.
func rummyTo[T usecase.Event](w *world, nickname string) []T {
	id := w.ids[nickname]
	var out []T
	for _, e := range w.broadcaster.Events() {
		if v, ok := e.Event.(T); ok && e.PlayerID == id && id != "" {
			out = append(out, v)
		}
	}
	return out
}

func lastTo[T usecase.Event](w *world, nickname string) (T, error) {
	all := rummyTo[T](w, nickname)
	if len(all) == 0 {
		var zero T
		return zero, fmt.Errorf("%q got no %T", nickname, zero)
	}
	return all[len(all)-1], nil
}

func (w *world) state(nickname string) (rummy.View, error) {
	s, err := lastTo[usecase.RummyState](w, nickname)
	return s.View, err
}

// rooms is the latest room list broadcast to the Praça.
func (w *world) rooms() ([]rummy.RoomInfo, error) {
	r, err := lastEvent[usecase.RummyRooms](w)
	return r.Rooms, err
}

func (w *world) roomOf(nickname string) (rummy.RoomInfo, error) {
	rooms, err := w.rooms()
	if err != nil {
		return rummy.RoomInfo{}, err
	}
	id := w.ids[nickname]
	for _, r := range rooms {
		if slices.ContainsFunc(r.Players, func(p rummy.Player) bool { return p.ID == id }) {
			return r, nil
		}
	}
	return rummy.RoomInfo{}, fmt.Errorf("%q is in no room", nickname)
}

// --- table and rooms ---

func (w *world) tableIsFree() error {
	if _, err := w.rooms(); err == nil {
		return fmt.Errorf("the table has rooms before the scenario")
	}
	return nil
}

// walkToTable walks down from the spawn point to just above the table.
func (w *world) walkToTable(nickname string) error {
	id, err := w.idOf(nickname)
	if err != nil {
		return err
	}
	if err := w.setInput.Execute(usecase.SetInputInput{PlayerID: id, Down: true}); err != nil {
		return err
	}
	near := false
	for range 200 {
		pos, err := w.position(nickname)
		if err != nil {
			return err
		}
		if err := w.runner.Exec(func(r *room.Room) { near = r.Map().NearTable(pos) }); err != nil {
			return err
		}
		if near {
			break
		}
		if err := w.advance(1); err != nil {
			return err
		}
	}
	if err := w.setInput.Execute(usecase.SetInputInput{PlayerID: id}); err != nil {
		return err
	}
	if !near {
		return fmt.Errorf("%q could not reach the table", nickname)
	}
	return nil
}

func (w *world) nearTable(nickname string) error {
	if err := w.ensureJoined(nickname); err != nil {
		return err
	}
	return w.walkToTable(nickname)
}

var modeWords = map[string]string{"Clássico": "classic", "Livre": "free"}

func (w *world) createRoom(nickname, mode string) error {
	m, ok := modeWords[mode]
	if !ok {
		return fmt.Errorf("unknown mode %q", mode)
	}
	return w.createTableRoom.Execute(w.ids[nickname], m)
}

func (w *world) createdRoom(nickname, mode string) error {
	if err := w.nearTable(nickname); err != nil {
		return err
	}
	return w.createRoom(nickname, mode)
}

func (w *world) lobbyWith(mode, nickname string) error {
	r, err := w.roomOf(nickname)
	if err != nil {
		return err
	}
	if r.Mode.String() != modeWords[mode] || r.Phase != rummy.Lobby || len(r.Players) != 1 {
		return fmt.Errorf("room = %+v, want a %s lobby with only %q", r, mode, nickname)
	}
	return nil
}

func (w *world) opensList(nickname string) error {
	if err := w.nearTable(nickname); err != nil {
		return err
	}
	return w.openTable.Execute(w.ids[nickname])
}

// listSeenBy is the room list last sent to a player who opened it.
func (w *world) listSeenBy(nickname, owner string) (rummy.RoomInfo, error) {
	list, err := lastTo[usecase.RummyRooms](w, nickname)
	if err != nil {
		return rummy.RoomInfo{}, err
	}
	for _, r := range list.Rooms {
		if slices.ContainsFunc(r.Players, func(p rummy.Player) bool { return p.Nickname == owner }) {
			return r, nil
		}
	}
	return rummy.RoomInfo{}, fmt.Errorf("%q does not see %q's room", nickname, owner)
}

func (w *world) seesRoomWithSeats(nickname, owner, mode string, seats int) error {
	r, err := w.listSeenBy(nickname, owner)
	if err != nil {
		return err
	}
	if r.Mode.String() != modeWords[mode] || rummy.MaxPlayers-len(r.Players) != seats {
		return fmt.Errorf("room = %+v, want %s with %d seats", r, mode, seats)
	}
	return nil
}

func (w *world) joinRoomOf(nickname, owner string) error {
	if err := w.nearTable(nickname); err != nil {
		return err
	}
	r, err := w.roomOf(owner)
	if err != nil {
		return err
	}
	return w.joinTableRoom.Execute(w.ids[nickname], r.ID)
}

func (w *world) joinsSameLobby(nickname, owner string) error {
	if err := w.joinRoomOf(nickname, owner); err != nil {
		return err
	}
	a, err := w.roomOf(nickname)
	if err != nil {
		return err
	}
	if b, _ := w.roomOf(owner); a.ID != b.ID || a.Phase != rummy.Lobby {
		return fmt.Errorf("%q and %q are not in the same lobby", nickname, owner)
	}
	return nil
}

func (w *world) countdownIn(seconds int, owner string) error {
	r, err := w.roomOf(owner)
	if err != nil {
		return err
	}
	if got := r.LobbyEnds.Sub(w.clock.Now()); got != time.Duration(seconds)*time.Second {
		return fmt.Errorf("countdown = %v, want %d s", got, seconds)
	}
	return nil
}

func (w *world) threeInLobby(a, b, c string) error {
	if err := w.createdRoom(a, "Livre"); err != nil {
		return err
	}
	for _, n := range []string{b, c} {
		if err := w.joinRoomOf(n, a); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) matchStartsWith(owner string, n int) error {
	r, err := w.roomOf(owner)
	if err != nil {
		return err
	}
	if r.Phase != rummy.Playing || len(r.Players) != n {
		return fmt.Errorf("room = %+v, want playing with %d", r, n)
	}
	return nil
}

// playTogether puts the players in a new Free room and runs the countdown.
func (w *world) playTogether(names ...string) error {
	if err := w.createdRoom(names[0], "Livre"); err != nil {
		return err
	}
	for _, n := range names[1:] {
		if err := w.joinRoomOf(n, names[0]); err != nil {
			return err
		}
	}
	if r, _ := w.roomOf(names[0]); r.Phase == rummy.Playing {
		return nil
	}
	return w.passTime(rummy.DefaultTiming().Lobby)
}

func (w *world) twoPlaying(a, b string) error        { return w.playTogether(a, b) }
func (w *world) threePlaying(a, b, c string) error   { return w.playTogether(a, b, c) }
func (w *world) fourPlaying(a, b, c, d string) error { return w.playTogether(a, b, c, d) }

func (w *world) twoMatches() error {
	s, err := lastEvent[usecase.RummyStatus](w)
	if err != nil {
		return err
	}
	if s.Matches != 2 {
		return fmt.Errorf("%d rooms playing, want 2", s.Matches)
	}
	return nil
}

func (w *world) noStateFrom(nickname, owner string) error {
	r, err := w.roomOf(owner)
	if err != nil {
		return err
	}
	for _, s := range rummyTo[usecase.RummyState](w, nickname) {
		if s.Room == r.ID {
			return fmt.Errorf("%q got the table of %q's room", nickname, owner)
		}
	}
	return nil
}

func (w *world) seesPlaying(nickname, owner string) error {
	r, err := w.listSeenBy(nickname, owner)
	if err != nil {
		return err
	}
	if r.Phase != rummy.Playing {
		return fmt.Errorf("room phase = %v, want playing", r.Phase)
	}
	return nil
}

func (w *world) cannotJoin(nickname, owner string) error {
	r, err := w.roomOf(owner)
	if err != nil {
		return err
	}
	if err := w.joinTableRoom.Execute(w.ids[nickname], r.ID); err != nil {
		return err
	}
	rej, err := lastTo[usecase.RummyRejected](w, nickname)
	if err != nil {
		return err
	}
	if rej.Reason != "playing" {
		return fmt.Errorf("rejected with %q, want playing", rej.Reason)
	}
	return nil
}

// players in the match a nickname is in, from their last state.
func (w *world) matchPlayers(nickname string) ([]string, error) {
	v, err := w.state(nickname)
	if err != nil {
		return nil, err
	}
	var out []string
	for nick, id := range w.ids {
		if _, ok := v.Counts[id]; ok {
			out = append(out, nick)
		}
	}
	return out, nil
}

func (w *world) everyHandHas(n int) error {
	for nick := range w.ids {
		v, err := w.state(nick)
		if err != nil {
			return err
		}
		if len(v.Hand) != n {
			return fmt.Errorf("%q has %d tiles, want %d", nick, len(v.Hand), n)
		}
	}
	return nil
}

func (w *world) poolHas(n int) error {
	for nick := range w.ids {
		v, err := w.state(nick)
		if err != nil {
			return err
		}
		if v.Pool != n {
			return fmt.Errorf("%q sees %d in the pool, want %d", nick, v.Pool, n)
		}
	}
	return nil
}

func (w *world) handIsSecret(viewer, other string) error {
	v, err := w.state(viewer)
	if err != nil {
		return err
	}
	o, err := w.state(other)
	if err != nil {
		return err
	}
	if v.Counts[w.ids[other]] != len(o.Hand) {
		return fmt.Errorf("%q sees %d tiles for %q, who has %d", viewer, v.Counts[w.ids[other]], other, len(o.Hand))
	}
	for _, s := range rummyTo[usecase.RummyState](w, viewer) {
		for _, id := range s.View.Hand {
			if slices.Contains(o.Hand, id) {
				return fmt.Errorf("%q received a tile of %q's hand", viewer, other)
			}
		}
	}
	return nil
}

// --- turns ---

// setUpTurn rigs the deal, starts a match of nickname against "bandit" and
// remembers the view at the start of nickname's turn.
func (w *world) setUpTurn(nickname, mode string, hand []string, opener []string) error {
	first := append(slices.Clone(opener), hand...)
	if err := w.rummy.shuffler.rig(first, nil); err != nil {
		return err
	}
	if err := w.createdRoom(nickname, mode); err != nil {
		return err
	}
	if err := w.joinRoomOf("bandit", nickname); err != nil {
		return err
	}
	if err := w.passTime(rummy.DefaultTiming().Lobby); err != nil {
		return err
	}
	if len(opener) > 0 {
		if err := w.layFaces(nickname, opener); err != nil {
			return err
		}
		if err := w.endTurn(nickname); err != nil {
			return err
		}
		if err := w.playTurn("bandit", usecase.ActionDraw, nil, nil); err != nil {
			return err
		}
	}
	return w.noteTurnStart(nickname)
}

func (w *world) noteTurnStart(nickname string) error {
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	if v.Current != w.ids[nickname] {
		return fmt.Errorf("it is not %q's turn", nickname)
	}
	w.rummy.start = v
	return nil
}

func (w *world) turnNotOpened(nickname, mode string) error {
	hand := []string{"blue-3", "blue-4", "blue-5", "red-10", "blue-10", "green-10"}
	return w.setUpTurn(nickname, mode, hand, nil)
}

func (w *world) tableHasRun(spec string) error {
	w.rummy.tableSeq = spec
	return nil
}

// opener is the meld played on the first turn so nickname has opened: the
// run the scenario put on the table, or blue 11-12-13.
func (w *world) opener(fill int) ([]string, error) {
	if w.rummy.tableSeq != "" {
		return faces(w.rummy.tableSeq)
	}
	if fill > 0 {
		var out []string
		for n := 1; n <= fill; n++ {
			out = append(out, fmt.Sprintf("brown-%d", n))
		}
		return out, nil
	}
	return []string{"blue-11", "blue-12", "blue-13"}, nil
}

func (w *world) turnOpenedWith(nickname, spec string) error {
	hand, err := faces(spec)
	if err != nil {
		return err
	}
	open, err := w.opener(0)
	if err != nil {
		return err
	}
	return w.setUpTurn(nickname, "Livre", hand, open)
}

func (w *world) turnOpenedOnly(nickname, spec string) error {
	hand, err := faces(spec)
	if err != nil {
		return err
	}
	open, err := w.opener(rummy.HandSize - len(hand))
	if err != nil {
		return err
	}
	return w.setUpTurn(nickname, "Livre", hand, open)
}

func (w *world) playTurn(nickname string, a usecase.TurnAction, melds [][]rummy.TileID, hand []rummy.TileID) error {
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	return w.playTurnOn(nickname, v.Turn, a, melds, hand)
}

func (w *world) playTurnOn(nickname string, turn int, a usecase.TurnAction, melds [][]rummy.TileID, hand []rummy.TileID) error {
	if a == usecase.ActionMove {
		w.rummy.lastMove.turn, w.rummy.lastMove.melds, w.rummy.lastMove.hand = turn, melds, hand
	}
	return w.playTurnUC.Execute(usecase.TurnInput{
		PlayerID: w.ids[nickname], Action: a, Turn: turn, Melds: melds, Hand: hand,
	})
}

// lay sends a move that adds a new meld from nickname's hand.
func (w *world) lay(nickname, spec string) error {
	want, err := faces(spec)
	if err != nil {
		return err
	}
	return w.layFaces(nickname, want)
}

func (w *world) layFaces(nickname string, want []string) error {
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	ids, rest, err := pick(v.Hand, want)
	if err != nil {
		return err
	}
	melds := append(slices.Clone(v.Table), ids)
	return w.playTurnOn(nickname, v.Turn, usecase.ActionMove, melds, rest)
}

func (w *world) endTurn(nickname string) error {
	return w.playTurn(nickname, usecase.ActionEnd, nil, nil)
}

func (w *world) layAndEnd(nickname, spec string) error {
	if err := w.lay(nickname, spec); err != nil {
		return err
	}
	return w.endTurn(nickname)
}

func (w *world) takeFromRun(nickname, tile, group string) error {
	taken, err := faces(tile)
	if err != nil {
		return err
	}
	want, err := faces(group)
	if err != nil {
		return err
	}
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	var from rummy.TileID
	melds := make([][]rummy.TileID, 0, len(v.Table)+1)
	found := false
	for _, m := range v.Table {
		if i := slices.IndexFunc(m, func(id rummy.TileID) bool { return faceOf(id) == taken[0] }); i >= 0 && !found {
			from, found = m[i], true
			m = slices.Delete(slices.Clone(m), i, i+1)
		}
		melds = append(melds, m)
	}
	if !found {
		return fmt.Errorf("no %s on the table", taken[0])
	}
	ids, rest, err := pick(append(slices.Clone(v.Hand), from), want)
	if err != nil {
		return err
	}
	melds = append(melds, ids)
	return w.playTurnOn(nickname, v.Turn, usecase.ActionMove, melds, rest)
}

func (w *world) undoTurn(nickname string) error {
	return w.playTurn(nickname, usecase.ActionUndo, nil, nil)
}

// lastTurn is the last turn result sent to nickname.
func (w *world) lastTurn(nickname string) (rummy.TurnEnded, error) {
	t, err := lastTo[usecase.RummyTurn](w, nickname)
	return t.Turn, err
}

// pluralColors names the colors as the client does, in the plural (RN12a).
var pluralColors = [rummy.Colors]string{"marrons", "vermelhos", "verdes", "azuis"}

// describe is the sentence the client shows for an issue (web
// src/game/rummy/issues.ts, describeIssue), for the codes the scenarios use.
func describe(is rummy.Issue) string {
	p := is.Params
	switch is.Code {
	case "opening_too_low":
		return fmt.Sprintf("Sua abertura soma %d pontos; faltam %d para 30.", p["points"], p["need"])
	case "too_short":
		return fmt.Sprintf("Esta combinação tem só %d peças; precisa de pelo menos 3.", p["count"])
	case "repeated_color":
		return fmt.Sprintf("O grupo de %d tem dois %d %s. Tire um deles.", p["value"], p["value"], pluralColors[p["color"]])
	}
	return is.Code
}

func (w *world) lastRefusal(nickname string) (rummy.Refused, error) {
	r, err := lastTo[usecase.RummyRefused](w, nickname)
	return r.Refused, err
}

func (w *world) playRefusedWith(nickname, reason string) error {
	r, err := w.lastRefusal(nickname)
	if err != nil {
		return err
	}
	var got []string
	for _, is := range r.Issues {
		if describe(is) == reason {
			return nil
		}
		got = append(got, describe(is))
	}
	return fmt.Errorf("refused with %q, want %q", got, reason)
}

func (w *world) tilesBlamed(spec string) error {
	want, err := faces(spec)
	if err != nil {
		return err
	}
	r, err := w.lastRefusal("rocky")
	if err != nil {
		return err
	}
	for _, is := range r.Issues {
		got := facesOf(is.Tiles)
		if len(got) == 0 && is.Meld >= 0 && is.Meld < len(w.rummy.lastMove.melds) {
			got = facesOf(w.rummy.lastMove.melds[is.Meld]) // the whole meld
		}
		if slices.Equal(got, want) {
			return nil
		}
	}
	return fmt.Errorf("issues %+v do not blame %v", r.Issues, want)
}

func (w *world) stillTurnOf(nickname string) error {
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	if v.Current != w.ids[nickname] || v.Turn != w.rummy.start.Turn {
		return fmt.Errorf("turn %d of %s, want turn %d of %q", v.Turn, v.Current, w.rummy.start.Turn, nickname)
	}
	return nil
}

func (w *world) didNotDraw(nickname string) error {
	if t, err := w.lastTurn(nickname); err == nil && t.Turn == w.rummy.start.Turn {
		return fmt.Errorf("turn ended: %+v", t)
	}
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	if got, want := v.Counts[w.ids[nickname]], len(w.rummy.start.Hand); got > want {
		return fmt.Errorf("%q has %d tiles, more than the %d at the start", nickname, got, want)
	}
	return nil
}

func (w *world) refusedPlay(spec, nickname string) error {
	if err := w.layAndEnd(nickname, spec); err != nil {
		return err
	}
	_, err := w.lastRefusal(nickname)
	return err
}

// addToLastAndEnd moves a tile from the hand to the end of the last meld the
// player laid. A refusal keeps the draft, so it starts from the last move.
func (w *world) addToLastAndEnd(nickname, tile string) error {
	want, err := faces(tile)
	if err != nil {
		return err
	}
	last := w.rummy.lastMove
	if len(last.melds) == 0 {
		return fmt.Errorf("%q has not moved yet", nickname)
	}
	melds := slices.Clone(last.melds)
	ids, rest, err := pick(last.hand, want)
	if err != nil {
		return err
	}
	n := len(melds) - 1
	melds[n] = append(slices.Clone(melds[n]), ids...)
	if err := w.playTurnOn(nickname, last.turn, usecase.ActionMove, melds, rest); err != nil {
		return err
	}
	return w.endTurn(nickname)
}

func (w *world) turnTimeRunsOut() error {
	return w.passTime(rummy.DefaultTiming().Turn)
}

func (w *world) timedOutAndDrewTile(nickname string) error {
	if err := w.timedOutAndDrew(nickname); err != nil {
		return err
	}
	return w.drewTile(nickname)
}

func (w *world) moveMeldTo(nickname, spec string, x, y int) error {
	want, err := faces(spec)
	if err != nil {
		return err
	}
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	pos := rummy.AutoPos(len(v.Table))
	i := slices.IndexFunc(v.Table, func(m []rummy.TileID) bool { return slices.Equal(facesOf(m), want) })
	if i < 0 {
		return fmt.Errorf("no %v on the table", want)
	}
	pos[i] = rummy.Pos{X: x, Y: y}
	return w.playTurnUC.Execute(usecase.TurnInput{
		PlayerID: w.ids[nickname], Action: usecase.ActionMove, Turn: v.Turn,
		Melds: v.Table, Pos: pos, Hand: v.Hand,
	})
}

func (w *world) seesMeldAt(nickname, spec string, x, y int) error {
	want, err := faces(spec)
	if err != nil {
		return err
	}
	d, err := lastTo[usecase.RummyDraft](w, nickname)
	if err != nil {
		return err
	}
	for i, m := range d.Melds {
		if slices.Equal(facesOf(m), want) {
			if d.Pos[i] != (rummy.Pos{X: x, Y: y}) {
				return fmt.Errorf("%v at %v, want %d, %d", want, d.Pos[i], x, y)
			}
			return nil
		}
	}
	return fmt.Errorf("draft %v has no %v", d.Melds, want)
}

func (w *world) tableSay(nickname, text string) error {
	return w.sendChat.Execute(usecase.SendChatInput{PlayerID: w.ids[nickname], Text: text, Table: true})
}

func (w *world) receivesTableChat(nickname, text, from string) error {
	c, err := lastTo[usecase.RummyChat](w, nickname)
	if err != nil {
		return err
	}
	if c.Text != text || c.From != w.ids[from] {
		return fmt.Errorf("table chat %+v, want %q from %q", c, text, from)
	}
	return nil
}

func (w *world) noTableChat(nickname string) error {
	if n := len(rummyTo[usecase.RummyChat](w, nickname)); n > 0 {
		return fmt.Errorf("%q got %d table chat lines", nickname, n)
	}
	for _, e := range w.broadcaster.Events() {
		if _, ok := e.Event.(usecase.RummyChat); ok && e.PlayerID == "" {
			return fmt.Errorf("a table chat line went to the whole room")
		}
	}
	return nil
}

func sameMelds(a, b [][]rummy.TileID) bool {
	return slices.EqualFunc(a, b, slices.Equal[[]rummy.TileID])
}

func (w *world) tableBackToStart() error {
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	if !sameMelds(v.Table, w.rummy.start.Table) {
		return fmt.Errorf("table = %v, want %v", v.Table, w.rummy.start.Table)
	}
	return nil
}

func (w *world) drewTile(nickname string) error {
	t, err := w.lastTurn(nickname)
	if err != nil {
		return err
	}
	if t.By != w.ids[nickname] || !t.Drew {
		return fmt.Errorf("last turn = %+v, want %q to draw", t, nickname)
	}
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	if len(v.Hand) != len(w.rummy.start.Hand)+1 {
		return fmt.Errorf("%q has %d tiles, want %d", nickname, len(v.Hand), len(w.rummy.start.Hand)+1)
	}
	return nil
}

func (w *world) playAccepted(nickname string) error {
	t, err := w.lastTurn(nickname)
	if err != nil {
		return err
	}
	if t.By != w.ids[nickname] || t.Result != rummy.ResultOK {
		return fmt.Errorf("last turn = %+v, want %q's play accepted", t, nickname)
	}
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	if len(v.Table) <= len(w.rummy.start.Table) {
		return fmt.Errorf("table = %v, want a new meld", v.Table)
	}
	return nil
}

func (w *world) nextPlayersTurn() error {
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	if v.Current != w.ids["bandit"] {
		return fmt.Errorf("turn of %s, want bandit", v.Current)
	}
	return nil
}

func (w *world) tableHolds(run, group string) error {
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	var want [][]string
	for _, spec := range []string{run, group} {
		f, err := faces(spec)
		if err != nil {
			return err
		}
		want = append(want, f)
	}
	var got [][]string
	for _, m := range v.Table {
		got = append(got, facesOf(m))
	}
	for _, f := range want {
		if !slices.ContainsFunc(got, func(g []string) bool { return slices.Equal(g, f) }) {
			return fmt.Errorf("table %v has no %v", got, f)
		}
	}
	if len(got) != len(want) {
		return fmt.Errorf("table %v, want only %v", got, want)
	}
	return nil
}

func (w *world) backToStart(nickname string) error {
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	if !sameMelds(v.Table, w.rummy.start.Table) || !slices.Equal(v.Hand, w.rummy.start.Hand) {
		return fmt.Errorf("table %v hand %v, want the start of the turn", v.Table, facesOf(v.Hand))
	}
	return nil
}

func (w *world) turnRunsOut(seconds int, _ string) error {
	return w.passTime(time.Duration(seconds) * time.Second)
}

func (w *world) timedOutAndDrew(nickname string) error {
	t, err := w.lastTurn(nickname)
	if err != nil {
		return err
	}
	if t.By != w.ids[nickname] || t.Result != rummy.ResultTimeout || !t.Drew {
		return fmt.Errorf("last turn = %+v, want %q timed out and drew", t, nickname)
	}
	return nil
}

func (w *world) results(nickname string) (rummy.MatchOver, error) {
	r, err := lastTo[usecase.RummyResults](w, nickname)
	return r.Results, err
}

func (w *world) everyoneSeesWinner(winner string) error {
	for nick := range w.ids {
		r, err := w.results(nick)
		if err != nil {
			return err
		}
		if r.Winner != w.ids[winner] {
			return fmt.Errorf("%q sees winner %s, want %q", nick, r.Winner, winner)
		}
	}
	return nil
}

func (w *world) scoreboardWithPoints() error {
	r, err := w.results("bandit")
	if err != nil {
		return err
	}
	if len(r.Scores) != 2 || r.Scores[0].Points != 0 || r.Scores[1].Points == 0 {
		return fmt.Errorf("scores = %+v, want the winner with 0 and the other with points", r.Scores)
	}
	return nil
}

// --- leaving ---

func (w *world) dropAndReturn(nickname string, seconds int) error {
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	w.rummy.startHands = map[string][]string{nickname: facesOf(v.Hand)}
	w.dropped[nickname] = true
	if err := w.passTime(time.Duration(seconds) * time.Second); err != nil {
		return err
	}
	delete(w.ids, nickname)
	delete(w.dropped, nickname)
	return w.mustJoin(nickname)
}

func (w *world) sameHandAfterReturn(nickname string) error {
	v, err := w.state(nickname)
	if err != nil {
		return err
	}
	got, want := facesOf(v.Hand), w.rummy.startHands[nickname]
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		return fmt.Errorf("hand = %v, want %v", got, want)
	}
	return nil
}

func (w *world) dropForGood(nickname string, seconds int) error {
	w.dropped[nickname] = true
	// Leaving the Praça takes room.PresenceTimeout; the seat is kept from then.
	return w.passTime(room.PresenceTimeout + time.Duration(seconds)*time.Second + time.Second)
}

func (w *world) removedTilesBack(nickname string) error {
	v, err := w.state("bandit")
	if err != nil {
		return err
	}
	if _, ok := v.Counts[w.ids[nickname]]; ok {
		return fmt.Errorf("%q is still in the match", nickname)
	}
	total := v.Pool
	for _, n := range v.Counts {
		total += n
	}
	for _, m := range v.Table {
		total += len(m)
	}
	if total != rummy.DeckSize {
		return fmt.Errorf("pool, hands and table hold %d tiles, want %d", total, rummy.DeckSize)
	}
	return nil
}

func (w *world) matchGoesOnForTwo(a, b string) error {
	players, err := w.matchPlayers(a)
	if err != nil {
		return err
	}
	slices.Sort(players)
	want := []string{a, b}
	slices.Sort(want)
	if !slices.Equal(players, want) {
		return fmt.Errorf("match players = %v, want %v", players, want)
	}
	if _, err := w.results(a); err == nil {
		return fmt.Errorf("the match ended")
	}
	return nil
}

func (w *world) leaveTable(nickname string) error {
	return w.leaveTableRoom.Execute(w.ids[nickname])
}

func (w *world) wins(nickname string) error {
	r, err := w.results(nickname)
	if err != nil {
		return err
	}
	if r.Winner != w.ids[nickname] {
		return fmt.Errorf("winner = %s, want %q", r.Winner, nickname)
	}
	return nil
}
