import {
  decodeServerMessage,
  encodeClientMessage,
  type Keys,
  type MeldPos,
  type RummyMode,
  type ServerMessage,
} from "./protocol";

export type Status = "connecting" | "open" | "reconnecting" | "closed";

export interface RoomClientOptions {
  onMessage: (msg: ServerMessage) => void;
  onStatus?: (status: Status) => void;
  /** Injected in tests; defaults to the browser WebSocket. */
  WebSocketImpl?: typeof WebSocket;
}

const OPEN = 1; // WebSocket.OPEN, without depending on the global in tests
const FIRST_RETRY_MS = 500;
const MAX_RETRY_MS = 5000;

/** Default server URL; override with NEXT_PUBLIC_WS_URL. */
export function defaultServerUrl(): string {
  return process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/ws";
}

/**
 * One connection to the room. join() connects; after an unexpected
 * disconnect it reconnects with exponential backoff (0.5s, 1s, 2s... up to
 * 5s) and joins again with the same nickname. The server assigns a new
 * player id on every connection, delivered in the next room.joined.
 *
 * Input and chat sent while disconnected are dropped: the server is
 * authoritative and stale input must not be replayed.
 */
export class RoomClient {
  private socket: WebSocket | null = null;
  private nickname: string | null = null;
  private seq = 0;
  private retryMs = FIRST_RETRY_MS;
  private retryTimer: ReturnType<typeof setTimeout> | null = null;
  private closed = false;

  constructor(
    private readonly url: string,
    private readonly opts: RoomClientOptions,
  ) {}

  join(nickname: string): void {
    this.nickname = nickname;
    this.closed = false;
    if (!this.socket) {
      this.connect("connecting");
    } else if (this.socket.readyState === OPEN) {
      this.sendJoin();
    }
  }

  sendInput(keys: Keys): void {
    this.seq++;
    this.send({ type: "player.input", data: { ...keys, seq: this.seq } });
  }

  sendChat(text: string): void {
    this.send({ type: "chat.send", data: { text } });
  }

  /** Three cups: E near the booth, a pick, and Esc / Sair (RW-005). */
  joinCups(): void {
    this.send({ type: "cups.join", data: {} });
  }

  pickCup(round: number, slot: number): void {
    this.send({ type: "cups.pick", data: { round, slot } });
  }

  leaveCups(): void {
    this.send({ type: "cups.leave", data: {} });
  }

  /** Rummy das Nozes (RW-006): the room list, rooms, turns and leaving. */
  rummyOpen(): void {
    this.send({ type: "rummy.open", data: {} });
  }

  rummyCreate(mode: RummyMode): void {
    this.send({ type: "rummy.create", data: { mode } });
  }

  rummyJoin(room: string): void {
    this.send({ type: "rummy.join", data: { room } });
  }

  rummyMove(turn: number, melds: number[][], pos: MeldPos[], hand: number[]): void {
    this.send({ type: "rummy.move", data: { turn, melds, pos, hand } });
  }

  rummyTurn(action: "end" | "draw" | "undo", turn: number): void {
    this.send({ type: `rummy.${action}`, data: { turn } });
  }

  rummyLeave(): void {
    this.send({ type: "rummy.leave", data: {} });
  }

  rummyAgain(): void {
    this.send({ type: "rummy.again", data: {} });
  }

  /** A line to the table chat, seen only by the match (RN37). */
  rummySay(text: string): void {
    this.send({ type: "rummy.say", data: { text } });
  }

  /** Closes for good: no reconnect. */
  close(): void {
    this.closed = true;
    if (this.retryTimer) clearTimeout(this.retryTimer);
    this.retryTimer = null;
    const socket = this.socket;
    this.socket = null;
    socket?.close();
    this.opts.onStatus?.("closed");
  }

  private connect(status: Status): void {
    this.opts.onStatus?.(status);
    const Impl = this.opts.WebSocketImpl ?? WebSocket;
    const socket = new Impl(this.url);
    this.socket = socket;

    socket.onopen = () => {
      this.retryMs = FIRST_RETRY_MS;
      this.opts.onStatus?.("open");
      this.sendJoin();
    };
    socket.onmessage = (e: MessageEvent) => {
      const msg = typeof e.data === "string" ? decodeServerMessage(e.data) : null;
      if (msg) this.opts.onMessage(msg);
    };
    socket.onclose = () => {
      if (this.socket !== socket || this.closed) return;
      this.socket = null;
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect(): void {
    this.opts.onStatus?.("reconnecting");
    const delay = this.retryMs;
    this.retryMs = Math.min(this.retryMs * 2, MAX_RETRY_MS);
    this.retryTimer = setTimeout(() => {
      this.retryTimer = null;
      if (!this.closed) this.connect("reconnecting");
    }, delay);
  }

  private sendJoin(): void {
    if (this.nickname !== null) {
      this.send({ type: "room.join", data: { nickname: this.nickname } });
    }
  }

  private send(msg: Parameters<typeof encodeClientMessage>[0]): void {
    if (this.socket?.readyState === OPEN) {
      this.socket.send(encodeClientMessage(msg));
    }
  }
}
