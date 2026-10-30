import type {
  MeldPos,
  PlayerId,
  RummyIssue,
  RummyMode,
  RummyPhase,
  RummyResultsData,
  RummyStateData,
  RummyTurnData,
  ServerMessage,
} from "@/net/protocol";

import { autoPos, type Layout } from "./layout";

/** A room of the table, with its countdown as a local time. */
export interface RummyRoom {
  id: string;
  mode: RummyMode;
  players: PlayerId[];
  phase: RummyPhase;
  /** When the countdown ends (performance.now), or null. */
  lobbyEndsAt: number | null;
}

/** Our match, as the server last sent it, plus the start of the turn. */
export interface RummyMatch extends RummyStateData {
  /** When the turn ends (performance.now). */
  turnEndsAt: number;
  /** How long the turn lasts in all, for the ring around the avatar (RN29). */
  turnTotal: number;
  /** The table and our hand when the turn started: Desfazer goes back here. */
  startMelds: number[][];
  startPos: MeldPos[];
  startHand: number[];
}

/** A line of the table chat (RN37), with when it arrived for its balloon (RN38). */
export interface TableChatLine {
  from: PlayerId;
  text: string;
  at: number;
}

/** How long a chat balloon stays next to an avatar (RN38). */
export const BALLOON_MS = 5000;
const CHAT_KEEP = 50;

/** What the client knows about the pieces table (RW-006). */
export interface RummyState {
  /** Who is at the table, waiting or playing (RN27). */
  playing: PlayerId[];
  /** Rooms playing a match (RN25). */
  tables: number;
  rooms: RummyRoom[];
  /** The room list is open: E near the table (RN5). Local only. */
  listOpen: boolean;
  match: RummyMatch | null;
  lastTurn: RummyTurnData | null;
  results: RummyResultsData | null;
  /** Why the last request was refused, until the next one. */
  rejection: string | null;
  /** Why our last play was refused, until we move again or the turn ends (RN12a). */
  refusal: RummyIssue[] | null;
  /** The table chat of this match (RN37). */
  chat: TableChatLine[];
}

export const initialRummyState: RummyState = {
  playing: [],
  tables: 0,
  rooms: [],
  listOpen: false,
  match: null,
  lastTurn: null,
  results: null,
  rejection: null,
  refusal: null,
  chat: [],
};

function roomsFrom(
  rooms: { id: string; mode: RummyMode; players: PlayerId[]; phase: RummyPhase; lobbyMs: number }[],
  at: number,
): RummyRoom[] {
  return rooms.map(({ lobbyMs, ...r }) => ({
    ...r,
    lobbyEndsAt: lobbyMs > 0 ? at + lobbyMs : null,
  }));
}

/**
 * Applies a server message received at `at` (performance.now) for player
 * `you`. Messages that are not about the table return the same object.
 */
export function rummyReducer(
  state: RummyState,
  msg: ServerMessage,
  at: number,
  you: PlayerId | null,
): RummyState {
  switch (msg.type) {
    case "room.joined": {
      const r = msg.data.rummy;
      if (!r) return initialRummyState;
      return {
        ...initialRummyState,
        playing: r.playing,
        tables: r.tables,
        rooms: roomsFrom(r.rooms, at),
      };
    }
    case "rummy.rooms": {
      const rooms = roomsFrom(msg.data.rooms, at);
      const mine = you === null ? undefined : rooms.find((r) => r.players.includes(you));
      // Back in the lobby after "Jogar de novo": the last match is over (RN7).
      if (mine?.phase === "lobby" && state.match) {
        return {
          ...state,
          rooms,
          match: null,
          results: null,
          lastTurn: null,
          refusal: null,
          chat: [],
        };
      }
      return { ...state, rooms };
    }
    case "rummy.status": {
      const next = { ...state, playing: msg.data.playing, tables: msg.data.tables };
      // Released from the table: the match and the scoreboard are over for us.
      if (you !== null && !msg.data.playing.includes(you)) {
        return { ...next, match: null, results: null, lastTurn: null, refusal: null, chat: [] };
      }
      return next;
    }
    case "rummy.rejected":
      return { ...state, rejection: msg.data.reason };
    case "rummy.state": {
      const d = msg.data;
      const prev = state.match;
      const newTurn = !prev || prev.room !== d.room || prev.turn !== d.turn;
      const sameRoom = prev !== null && prev.room === d.room;
      const pos = d.pos?.length === d.melds.length ? d.pos : autoPos(d.melds.length);
      return {
        ...state,
        listOpen: false,
        results: sameRoom ? state.results : null,
        rejection: null,
        refusal: newTurn ? null : state.refusal,
        chat: sameRoom ? state.chat : [],
        match: {
          ...d,
          pos,
          turnEndsAt: at + d.turnMs,
          turnTotal: newTurn ? d.turnMs : prev.turnTotal,
          startMelds: newTurn ? d.melds : prev.startMelds,
          startPos: newTurn ? pos : prev.startPos,
          startHand: newTurn ? d.hand : prev.startHand,
        },
      };
    }
    case "rummy.draft": {
      const m = state.match;
      if (!m || m.turn !== msg.data.turn || m.current === you) return state;
      return {
        ...state,
        match: {
          ...m,
          melds: msg.data.melds,
          pos:
            msg.data.pos?.length === msg.data.melds.length
              ? msg.data.pos
              : autoPos(msg.data.melds.length),
          counts: { ...m.counts, [m.current]: msg.data.handCount },
        },
      };
    }
    case "rummy.turn":
      return { ...state, lastTurn: msg.data };
    case "rummy.results":
      return { ...state, results: msg.data };
    case "rummy.refused":
      if (!state.match || state.match.turn !== msg.data.turn) return state;
      return { ...state, refusal: msg.data.issues };
    case "rummy.chat":
      return { ...state, chat: [...state.chat, { ...msg.data, at }].slice(-CHAT_KEEP) };
    default:
      return state;
  }
}

/** Our own move, shown at once while it goes to the server (RN11). */
export function moveLocally(state: RummyState, layout: Layout): RummyState {
  if (!state.match) return state;
  const { melds, pos, hand } = layout;
  return { ...state, refusal: null, match: { ...state.match, melds, pos, hand } };
}

/** Opens or closes the room list. */
export function setListOpen(state: RummyState, open: boolean): RummyState {
  return { ...state, listOpen: open, rejection: null };
}

/** The room `id` is in, if any. */
export function roomOf(state: RummyState, id: PlayerId | null): RummyRoom | null {
  if (id === null) return null;
  return state.rooms.find((r) => r.players.includes(id)) ?? null;
}

/**
 * Whole seconds left until `endsAt`. The frame time stops in a background
 * tab, so the clock is read too, to never show more than the server gave.
 */
export function secondsUntil(endsAt: number, frameNow: number): number {
  const now = Math.max(frameNow, performance.now());
  return Math.max(0, Math.ceil((endsAt - now) / 1000));
}
