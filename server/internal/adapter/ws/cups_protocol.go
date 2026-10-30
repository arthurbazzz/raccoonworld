package ws

import (
	"time"

	"raccoons-world/server/internal/domain/cups"
	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
)

// Three cups message types (see specs/005-tres-copos/plan.md, section 3.4).
// They are additions to protocol v1: older clients ignore unknown types.
const (
	TypeCupsJoin  = "cups.join"
	TypeCupsPick  = "cups.pick"
	TypeCupsLeave = "cups.leave"

	TypeCupsStatus   = "cups.status"
	TypeCupsRejected = "cups.rejected"
	TypeCupsRound    = "cups.round"
	TypeCupsPicked   = "cups.picked"
	TypeCupsReveal   = "cups.reveal"
	TypeCupsResults  = "cups.results"
)

// CupsJoinRequest is "cups.join": press E near the booth.
type CupsJoinRequest struct{}

// CupsPickRequest is "cups.pick": a cup slot (0-2, left to right) for a round.
type CupsPickRequest struct {
	Round int `json:"round"`
	Slot  int `json:"slot"`
}

// CupsLeaveRequest is "cups.leave": Esc or the Sair button.
type CupsLeaveRequest struct{}

func (CupsJoinRequest) isClientMessage()  {}
func (CupsPickRequest) isClientMessage()  {}
func (CupsLeaveRequest) isClientMessage() {}

// ToInput builds the PickCup input for the connection's player ID.
func (m CupsPickRequest) ToInput(id room.PlayerID) usecase.PickInput {
	return usecase.PickInput{PlayerID: id, Round: m.Round, Slot: m.Slot}
}

// Wire shapes of the three cups data.
type (
	cupsStatusJSON struct {
		Phase   string          `json:"phase"`
		Players []room.PlayerID `json:"players"`
		Queued  []room.PlayerID `json:"queued"`
		LobbyMs int64           `json:"lobbyMs"`
	}
	cupsRoundJSON struct {
		Round     int      `json:"round"`
		Total     int      `json:"total"`
		Nut       int      `json:"nut"`
		ShowMs    int64    `json:"showMs"`
		Swaps     [][2]int `json:"swaps"`
		SwapMs    int64    `json:"swapMs"`
		PickMs    int64    `json:"pickMs"`
		ElapsedMs int64    `json:"elapsedMs"`
	}
	cupsSnapshotJSON struct {
		cupsStatusJSON
		Round *cupsRoundJSON `json:"round"`
	}
	cupsPickJSON struct {
		ID   room.PlayerID `json:"id"`
		Slot *int          `json:"slot"` // null: no pick in time
		Hit  bool          `json:"hit"`
	}
	cupsRevealJSON struct {
		Round int            `json:"round"`
		Nut   int            `json:"nut"`
		Picks []cupsPickJSON `json:"picks"`
	}
	cupsScoreJSON struct {
		ID    room.PlayerID `json:"id"`
		Hits  int           `json:"hits"`
		Place int           `json:"place"`
	}
	cupsResultsJSON struct {
		Scores []cupsScoreJSON `json:"scores"`
	}
	cupsRejectedJSON struct {
		Reason usecase.CupsRejectReason `json:"reason"`
	}
)

func phaseName(p cups.Phase) string {
	switch {
	case p == cups.Lobby:
		return "lobby"
	case p.Playing():
		return "playing"
	case p == cups.Results:
		return "results"
	default:
		return "idle"
	}
}

func ms(d time.Duration) int64 { return d.Milliseconds() }

func toCupsStatusJSON(s cups.StatusChanged) cupsStatusJSON {
	out := cupsStatusJSON{Phase: phaseName(s.Phase), Players: s.Players, Queued: s.Queued, LobbyMs: ms(s.LobbyLeft)}
	if out.Players == nil {
		out.Players = []room.PlayerID{}
	}
	if out.Queued == nil {
		out.Queued = []room.PlayerID{}
	}
	return out
}

func toCupsRoundJSON(r usecase.CupsRound) cupsRoundJSON {
	swaps := make([][2]int, len(r.Round.Swaps))
	for i, s := range r.Round.Swaps {
		swaps[i] = [2]int(s)
	}
	return cupsRoundJSON{
		Round: r.Round.Round, Total: r.Round.Total, Nut: r.Round.Nut,
		ShowMs: ms(r.Round.Show), Swaps: swaps, SwapMs: ms(r.Round.SwapTime),
		PickMs: ms(r.Round.Pick), ElapsedMs: ms(r.Elapsed),
	}
}

func toCupsSnapshotJSON(s usecase.CupsSnapshot) cupsSnapshotJSON {
	out := cupsSnapshotJSON{cupsStatusJSON: toCupsStatusJSON(s.Status)}
	if s.Round != nil {
		r := toCupsRoundJSON(*s.Round)
		out.Round = &r
	}
	return out
}

// cupsWireData encodes the three cups events; ok is false for other events.
func cupsWireData(ev usecase.Event) (typ string, data any, ok bool) {
	switch e := ev.(type) {
	case usecase.CupsStatus:
		return TypeCupsStatus, toCupsStatusJSON(e.Status), true
	case usecase.CupsRejected:
		return TypeCupsRejected, cupsRejectedJSON{Reason: e.Reason}, true
	case usecase.CupsRound:
		return TypeCupsRound, toCupsRoundJSON(e), true
	case usecase.CupsPicked:
		return TypeCupsPicked, leftJSON{ID: e.ID}, true
	case usecase.CupsRevealed:
		picks := make([]cupsPickJSON, len(e.Reveal.Picks))
		for i, p := range e.Reveal.Picks {
			picks[i] = cupsPickJSON{ID: p.ID, Hit: p.Hit}
			if p.Picked {
				slot := p.Slot
				picks[i].Slot = &slot
			}
		}
		return TypeCupsReveal, cupsRevealJSON{Round: e.Reveal.Round, Nut: e.Reveal.Nut, Picks: picks}, true
	case usecase.CupsResults:
		scores := make([]cupsScoreJSON, len(e.Results.Scores))
		for i, s := range e.Results.Scores {
			scores[i] = cupsScoreJSON{ID: s.ID, Hits: s.Hits, Place: s.Place}
		}
		return TypeCupsResults, cupsResultsJSON{Scores: scores}, true
	}
	return "", nil, false
}
