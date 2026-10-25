import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { RoomClient, type Status } from "./client";
import type { ServerMessage } from "./protocol";

/** Minimal WebSocket double: tests drive open/message/close by hand. */
class FakeSocket {
  static instances: FakeSocket[] = [];
  static readonly OPEN = 1;

  readyState = 0;
  sent: unknown[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;

  constructor(public url: string) {
    FakeSocket.instances.push(this);
  }

  send(raw: string) {
    this.sent.push(JSON.parse(raw));
  }

  close() {
    this.readyState = 3;
    this.onclose?.();
  }

  // Test helpers.
  open() {
    this.readyState = FakeSocket.OPEN;
    this.onopen?.();
  }
  receive(msg: object) {
    this.onmessage?.({ data: JSON.stringify({ v: 1, ...msg }) });
  }
  drop() {
    this.readyState = 3;
    this.onclose?.();
  }
}

const latest = () => FakeSocket.instances[FakeSocket.instances.length - 1];

function newClient() {
  const messages: ServerMessage[] = [];
  const statuses: Status[] = [];
  const client = new RoomClient("ws://test/ws", {
    WebSocketImpl: FakeSocket as unknown as typeof WebSocket,
    onMessage: (m) => messages.push(m),
    onStatus: (s) => statuses.push(s),
  });
  return { client, messages, statuses };
}

beforeEach(() => {
  FakeSocket.instances = [];
  vi.useFakeTimers();
});
afterEach(() => vi.useRealTimers());

describe("RoomClient", () => {
  it("joins once the socket is open", () => {
    const { client, statuses } = newClient();
    client.join("rocky");
    expect(latest().sent).toEqual([]); // not open yet

    latest().open();

    expect(latest().sent).toEqual([{ type: "room.join", v: 1, data: { nickname: "rocky" } }]);
    expect(statuses).toEqual(["connecting", "open"]);
  });

  it("delivers decoded server messages and skips invalid ones", () => {
    const { client, messages } = newClient();
    client.join("rocky");
    latest().open();

    latest().onmessage?.({ data: "{" });
    latest().receive({ type: "player.left", data: { id: "c9" } });

    expect(messages).toEqual([{ type: "player.left", data: { id: "c9" } }]);
  });

  it("numbers input messages", () => {
    const { client } = newClient();
    client.join("rocky");
    latest().open();

    client.sendInput({ up: false, down: false, left: false, right: true });
    client.sendInput({ up: false, down: false, left: false, right: false });

    const inputs = latest().sent.filter((m) => (m as { type: string }).type === "player.input");
    expect(inputs).toEqual([
      {
        type: "player.input",
        v: 1,
        data: { up: false, down: false, left: false, right: true, seq: 1 },
      },
      {
        type: "player.input",
        v: 1,
        data: { up: false, down: false, left: false, right: false, seq: 2 },
      },
    ]);
  });

  it("drops input and chat while disconnected", () => {
    const { client } = newClient();
    client.join("rocky");
    client.sendChat("oi");
    client.sendInput({ up: true, down: false, left: false, right: false });
    latest().open();

    expect(latest().sent).toEqual([{ type: "room.join", v: 1, data: { nickname: "rocky" } }]);
  });

  it("reconnects with backoff and joins again with the same nickname", () => {
    const { client, statuses } = newClient();
    client.join("rocky");
    latest().open();

    latest().drop();
    expect(statuses.at(-1)).toBe("reconnecting");
    expect(FakeSocket.instances).toHaveLength(1);

    vi.advanceTimersByTime(500);
    expect(FakeSocket.instances).toHaveLength(2);
    latest().open();

    expect(latest().sent).toEqual([{ type: "room.join", v: 1, data: { nickname: "rocky" } }]);
    expect(statuses.at(-1)).toBe("open");
  });

  it("backs off longer after repeated failures, up to a cap", () => {
    const { client } = newClient();
    client.join("rocky");
    latest().drop(); // 500ms
    vi.advanceTimersByTime(500);
    latest().drop(); // 1000ms
    vi.advanceTimersByTime(999);
    expect(FakeSocket.instances).toHaveLength(2);
    vi.advanceTimersByTime(1);
    expect(FakeSocket.instances).toHaveLength(3);

    for (let i = 0; i < 10; i++) {
      latest().drop();
      vi.advanceTimersByTime(5000);
    }
    expect(FakeSocket.instances).toHaveLength(13); // every retry within the 5s cap
  });

  it("does not reconnect after close()", () => {
    const { client, statuses } = newClient();
    client.join("rocky");
    latest().open();

    client.close();
    vi.advanceTimersByTime(10_000);

    expect(FakeSocket.instances).toHaveLength(1);
    expect(statuses.at(-1)).toBe("closed");
  });
});
