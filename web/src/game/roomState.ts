import type { PlayerId, PlayerState, RejectReason, ServerMessage } from "@/net/protocol";

export type Phase = "nickname" | "joining" | "playing";

export interface ChatLine {
  from: PlayerId;
  nickname: string;
  text: string;
}

/** How many chat lines are kept for the accessible message log. */
const MAX_CHAT_LINES = 20;

/** What the client knows about the room, built only from server messages. */
export interface RoomState {
  phase: Phase;
  nickname: string;
  you: PlayerId | null;
  map: string | null;
  players: Record<PlayerId, PlayerState>;
  tick: number;
  rejection: RejectReason | null;
  messages: ChatLine[];
}

export const initialRoomState: RoomState = {
  phase: "nickname",
  nickname: "",
  you: null,
  map: null,
  players: {},
  tick: 0,
  rejection: null,
  messages: [],
};

/** The player submitted a nickname and is waiting for the server. */
export function startJoin(state: RoomState, nickname: string): RoomState {
  return { ...state, phase: "joining", nickname, rejection: null };
}

/** Applies one server message. Returns the same object when nothing changes. */
export function roomReducer(state: RoomState, msg: ServerMessage): RoomState {
  switch (msg.type) {
    case "room.joined": {
      const players: Record<PlayerId, PlayerState> = {};
      for (const p of msg.data.players) players[p.id] = p;
      return {
        ...state,
        phase: "playing",
        you: msg.data.you,
        map: msg.data.map,
        players,
        rejection: null,
        messages: [],
      };
    }
    case "room.join_rejected":
      // A duplicate join from a reconnect race: we are still in the room.
      if (msg.data.reason === "already_joined") return state;
      return { ...initialRoomState, nickname: state.nickname, rejection: msg.data.reason };
    default:
      if (state.phase !== "playing") return state;
      return applyRoomEvent(state, msg);
  }
}

function applyRoomEvent(state: RoomState, msg: ServerMessage): RoomState {
  switch (msg.type) {
    case "player.joined":
      if (msg.data.id === state.you) return state; // our own arrival
      return { ...state, players: { ...state.players, [msg.data.id]: msg.data } };
    case "player.left": {
      if (!(msg.data.id in state.players)) return state;
      const players = { ...state.players };
      delete players[msg.data.id];
      return { ...state, players };
    }
    case "room.snapshot": {
      const players = { ...state.players };
      for (const p of msg.data.players) {
        const known = players[p.id];
        if (known) players[p.id] = { ...known, x: p.x, y: p.y };
      }
      return { ...state, players, tick: msg.data.tick };
    }
    case "chat.message": {
      const sender = state.players[msg.data.from];
      if (!sender) return state;
      const line = { from: msg.data.from, nickname: sender.nickname, text: msg.data.text };
      return { ...state, messages: [...state.messages, line].slice(-MAX_CHAT_LINES) };
    }
    default:
      return state;
  }
}
