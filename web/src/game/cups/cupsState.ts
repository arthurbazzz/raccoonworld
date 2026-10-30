import type {
  CupsPhase,
  CupsRevealData,
  CupsRoundData,
  CupsScore,
  PlayerId,
  ServerMessage,
} from "@/net/protocol";

/** A round as the client animates it: started `startedAt` (performance.now). */
export interface CupsRound extends CupsRoundData {
  startedAt: number;
}

/** What the client knows about the three cups booth (RW-005). */
export interface CupsState {
  phase: CupsPhase;
  /** In the lobby or the match, in join order. */
  players: PlayerId[];
  /** Waiting for the next match (RC5). */
  queued: PlayerId[];
  /** When the lobby countdown ends (performance.now), in the lobby. */
  lobbyEndsAt: number | null;
  round: CupsRound | null;
  /** Who picked in this round; the cup stays hidden until the reveal (RC10). */
  picked: PlayerId[];
  /** Our own pick in this round, kept locally as we send it. */
  myPick: number | null;
  reveal: CupsRevealData | null;
  scores: CupsScore[] | null;
  /** The last E was too far from the booth. */
  tooFar: boolean;
}

export const initialCupsState: CupsState = {
  phase: "idle",
  players: [],
  queued: [],
  lobbyEndsAt: null,
  round: null,
  picked: [],
  myPick: null,
  reveal: null,
  scores: null,
  tooFar: false,
};

function withRound(state: CupsState, round: CupsRoundData, at: number): CupsState {
  return {
    ...state,
    round: { ...round, startedAt: at - round.elapsedMs },
    picked: [],
    myPick: null,
    reveal: null,
  };
}

/**
 * Applies a server message received at `at` (performance.now). Messages
 * that are not about the booth return the same object.
 */
export function cupsReducer(state: CupsState, msg: ServerMessage, at: number): CupsState {
  switch (msg.type) {
    case "room.joined": {
      const cups = msg.data.cups;
      if (!cups) return initialCupsState;
      const next = applyStatus(
        initialCupsState,
        cups.phase,
        cups.players,
        cups.queued,
        cups.lobbyMs,
        at,
      );
      return cups.round ? withRound(next, cups.round, at) : next;
    }
    case "cups.status": {
      const { phase, players, queued, lobbyMs } = msg.data;
      return applyStatus(state, phase, players, queued, lobbyMs, at);
    }
    case "cups.rejected":
      return { ...state, tooFar: true };
    case "cups.round":
      return withRound(state, msg.data, at);
    case "cups.picked":
      if (state.picked.includes(msg.data.id)) return state;
      return { ...state, picked: [...state.picked, msg.data.id] };
    case "cups.reveal":
      return { ...state, reveal: msg.data };
    case "cups.results":
      return { ...state, scores: msg.data.scores };
    default:
      return state;
  }
}

function applyStatus(
  state: CupsState,
  phase: CupsPhase,
  players: PlayerId[],
  queued: PlayerId[],
  lobbyMs: number,
  at: number,
): CupsState {
  const next: CupsState = {
    ...state,
    phase,
    players,
    queued,
    lobbyEndsAt: phase === "lobby" ? at + lobbyMs : null,
    tooFar: false,
  };
  // A new lobby, or a free booth, forgets the last match.
  if (phase === "lobby" || phase === "idle") {
    return { ...next, round: null, picked: [], myPick: null, reveal: null, scores: null };
  }
  return next;
}

/** Records our own pick as we send it; only the first one counts (RC9). */
export function pickLocally(state: CupsState, slot: number): CupsState {
  if (state.myPick !== null) return state;
  return { ...state, myPick: slot };
}

/** Where `id` stands at the booth. */
export function cupsRole(state: CupsState, id: PlayerId | null): "player" | "queued" | null {
  if (id === null) return null;
  if (state.players.includes(id)) return "player";
  if (state.queued.includes(id)) return "queued";
  return null;
}
