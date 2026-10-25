// WebSocket protocol v1 (specs/001-sala-inicial/plan.md, section 3.4).
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
  | { type: "chat.send"; data: { text: string } };

export interface PlayerState {
  id: PlayerId;
  nickname: string;
  x: number;
  y: number;
}

export type RejectReason = "nickname_taken" | "nickname_invalid" | "room_full" | "already_joined";

// Server -> client.
export type ServerMessage =
  | { type: "room.joined"; data: { you: PlayerId; map: string; players: PlayerState[] } }
  | { type: "room.join_rejected"; data: { reason: RejectReason } }
  | { type: "player.joined"; data: PlayerState }
  | { type: "player.left"; data: { id: PlayerId } }
  | {
      type: "room.snapshot";
      data: { tick: number; players: { id: PlayerId; x: number; y: number }[] };
    }
  | { type: "chat.message"; data: { from: PlayerId; text: string } };

const serverTypes: ReadonlySet<string> = new Set<ServerMessage["type"]>([
  "room.joined",
  "room.join_rejected",
  "player.joined",
  "player.left",
  "room.snapshot",
  "chat.message",
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
