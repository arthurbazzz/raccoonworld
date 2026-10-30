// WebSocket protocol v1 (specs/001-sala-inicial/plan.md, section 3.4, and
// specs/005-tres-copos/plan.md, section 3.4, and specs/006-rummy/plan.md,
// section 3.5).
// Mirrors server/internal/adapter/ws/protocol.go.

export const PROTOCOL_VERSION = 1;

export type PlayerId = string;

export interface Keys {
  up: boolean;
  down: boolean;
  left: boolean;
  right: boolean;
}

// Client -> server.
export type ClientMessage =
  | { type: "room.join"; data: { nickname: string } }
  | { type: "player.input"; data: Keys & { seq: number } }
  | { type: "chat.send"; data: { text: string } }
  | { type: "cups.join"; data: Record<string, never> }
  | { type: "cups.pick"; data: { round: number; slot: number } }
  | { type: "rummy.open"; data: Record<string, never> }
  | { type: "rummy.create"; data: { mode: RummyMode } }
  | { type: "rummy.join"; data: { room: string } }
  | { type: "rummy.move"; data: { turn: number; melds: number[][]; hand: number[] } }
  | { type: "rummy.end" | "rummy.draw" | "rummy.undo"; data: { turn: number } }
  | { type: "rummy.leave"; data: Record<string, never> }
  | { type: "rummy.again"; data: Record<string, never> }
  | { type: "cups.leave"; data: Record<string, never> };

export interface PlayerState {
  id: PlayerId;
  nickname: string;
  x: number;
  y: number;
}

// Three cups (specs/005-tres-copos/plan.md, section 3.4). Slots are 0, 1, 2
// from left to right; players see them as cups 1, 2 and 3.

export type CupsPhase = "idle" | "lobby" | "playing" | "results";

export interface CupsStatusData {
  phase: CupsPhase;
  players: PlayerId[];
  queued: PlayerId[];
  /** Time left in the lobby when the message was sent. */
  lobbyMs: number;
}

export interface CupsRoundData {
  round: number;
  total: number;
  /** Slot the nut starts under. */
  nut: number;
  showMs: number;
  swaps: [number, number][];
  swapMs: number;
  pickMs: number;
  /** How long ago the round started (only for newcomers, in room.joined). */
  elapsedMs: number;
}

export interface CupsRevealData {
  round: number;
  nut: number;
  /** slot is null for whoever did not pick in time. */
  picks: { id: PlayerId; slot: number | null; hit: boolean }[];
}

export interface CupsScore {
  id: PlayerId;
  hits: number;
  /** Ties share a place. */
  place: number;
}

// Rummy das Nozes (specs/006-rummy/plan.md, section 3.5). Tiles travel as
// IDs 0-105; game/rummy/tiles.ts says what each one is.

export type RummyMode = "classic" | "free";
export type RummyPhase = "lobby" | "playing" | "results";

export interface RummyRoomData {
  id: string;
  mode: RummyMode;
  players: PlayerId[];
  phase: RummyPhase;
  /** Countdown left when the message was sent; 0 without one. */
  lobbyMs: number;
}

export interface RummyStatusData {
  playing: PlayerId[];
  /** Rooms playing a match. */
  tables: number;
}

/** One player's view of their match; only their own hand (RN9). */
export interface RummyStateData {
  room: string;
  mode: RummyMode;
  turn: number;
  current: PlayerId;
  /** Time left in the turn when the message was sent. */
  turnMs: number;
  melds: number[][];
  hand: number[];
  counts: Record<PlayerId, number>;
  pool: number;
  opened: PlayerId[];
  absent: PlayerId[];
}

export type RummyResult = "ok" | "invalid" | "drew" | "timeout";

export interface RummyTurnData {
  turn: number;
  result: RummyResult;
  /** Why a play was refused (a meld error or an opening rule), or "". */
  reason: string;
  by: PlayerId;
  drew: boolean;
}

export interface RummyResultsData {
  /** null on a tie. */
  winner: PlayerId | null;
  scores: { id: PlayerId; points: number; place: number }[];
}

export type RejectReason = "nickname_taken" | "nickname_invalid" | "room_full" | "already_joined";

// Server -> client.
export type ServerMessage =
  | {
      type: "room.joined";
      data: {
        you: PlayerId;
        map: string;
        players: PlayerState[];
        /** The three cups booth right now (RW-005). */
        cups?: CupsStatusData & { round: CupsRoundData | null };
        /** The pieces table right now (RW-006). */
        rummy?: RummyStatusData & { rooms: RummyRoomData[] };
      };
    }
  | { type: "room.join_rejected"; data: { reason: RejectReason } }
  | { type: "player.joined"; data: PlayerState }
  | { type: "player.left"; data: { id: PlayerId } }
  | {
      type: "room.snapshot";
      data: { tick: number; players: { id: PlayerId; x: number; y: number }[] };
    }
  | { type: "chat.message"; data: { from: PlayerId; text: string } }
  | { type: "cups.status"; data: CupsStatusData }
  | { type: "cups.rejected"; data: { reason: "too_far" } }
  | { type: "cups.round"; data: CupsRoundData }
  | { type: "cups.picked"; data: { id: PlayerId } }
  | { type: "cups.reveal"; data: CupsRevealData }
  | { type: "cups.results"; data: { scores: CupsScore[] } }
  | { type: "rummy.rooms"; data: { rooms: RummyRoomData[] } }
  | { type: "rummy.status"; data: RummyStatusData }
  | { type: "rummy.rejected"; data: { reason: string } }
  | { type: "rummy.state"; data: RummyStateData }
  | { type: "rummy.draft"; data: { turn: number; melds: number[][]; handCount: number } }
  | { type: "rummy.turn"; data: RummyTurnData }
  | { type: "rummy.results"; data: RummyResultsData };

const serverTypes: ReadonlySet<string> = new Set<ServerMessage["type"]>([
  "room.joined",
  "room.join_rejected",
  "player.joined",
  "player.left",
  "room.snapshot",
  "chat.message",
  "cups.status",
  "cups.rejected",
  "cups.round",
  "cups.picked",
  "cups.reveal",
  "cups.results",
  "rummy.rooms",
  "rummy.status",
  "rummy.rejected",
  "rummy.state",
  "rummy.draft",
  "rummy.turn",
  "rummy.results",
]);

export function encodeClientMessage(msg: ClientMessage): string {
  return JSON.stringify({ type: msg.type, v: PROTOCOL_VERSION, data: msg.data });
}

/**
 * Parses a server message. Returns null for anything that is not a v1
 * server-to-client envelope; the payload shape is trusted, as the server
 * is authoritative and its output is covered by golden tests.
 */
export function decodeServerMessage(raw: string): ServerMessage | null {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof parsed !== "object" || parsed === null) return null;
  const { type, v, data } = parsed as { type?: unknown; v?: unknown; data?: unknown };
  if (typeof type !== "string" || !serverTypes.has(type)) return null;
  if (v !== PROTOCOL_VERSION) return null;
  if (typeof data !== "object" || data === null) return null;
  return { type, data } as ServerMessage;
}
