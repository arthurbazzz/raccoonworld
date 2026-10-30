import { describe, expect, it } from "vitest";

import type { RummyStateData, ServerMessage } from "@/net/protocol";

import { initialRummyState, moveLocally, roomOf, rummyReducer, setListOpen } from "./rummyState";

const state = (over: Partial<RummyStateData> = {}): ServerMessage => ({
  type: "rummy.state",
  data: {
    room: "s1",
    mode: "free",
    turn: 3,
    current: "p1",
    turnMs: 40000,
    melds: [[1, 2, 3]],
    hand: [10, 11],
    counts: { p1: 2, p2: 14 },
    pool: 75,
    opened: [],
    absent: [],
    ...over,
  },
});

describe("rummyReducer", () => {
  it("reads the table from room.joined", () => {
    const s = rummyReducer(
      initialRummyState,
      {
        type: "room.joined",
        data: {
          you: "p1",
          map: "praca",
          players: [],
          rummy: {
            playing: ["p2"],
            tables: 1,
            rooms: [{ id: "s1", mode: "classic", players: ["p2"], phase: "lobby", lobbyMs: 5000 }],
          },
        },
      },
      100,
      "p1",
    );
    expect(s.playing).toEqual(["p2"]);
    expect(s.tables).toBe(1);
    expect(s.rooms[0].lobbyEndsAt).toBe(5100);
  });

  it("keeps the start of the turn while moves come in, and resets it on a new turn", () => {
    let s = rummyReducer(initialRummyState, state(), 0, "p1");
    expect(s.match?.turnEndsAt).toBe(40000);
    s = moveLocally(s, [[1, 2, 3], [10]], [11]);
    expect(s.match?.hand).toEqual([11]);
    expect(s.match?.startHand).toEqual([10, 11]);

    s = rummyReducer(s, state({ turn: 4, current: "p2", hand: [11, 12] }), 0, "p1");
    expect(s.match?.startHand).toEqual([11, 12]);
  });

  it("shows the others' draft live, but not over our own turn", () => {
    const draft: ServerMessage = {
      type: "rummy.draft",
      data: { turn: 3, melds: [[1, 2, 3], [7, 8, 9]], handCount: 11 },
    };
    const watching = rummyReducer(rummyReducer(initialRummyState, state(), 0, "p2"), draft, 0, "p2");
    expect(watching.match?.melds).toHaveLength(2);
    expect(watching.match?.counts.p1).toBe(11);

    const mine = rummyReducer(initialRummyState, state(), 0, "p1");
    expect(rummyReducer(mine, draft, 0, "p1")).toBe(mine);
  });

  it("forgets the match once we are released from the table", () => {
    let s = rummyReducer(initialRummyState, state(), 0, "p1");
    s = rummyReducer(s, { type: "rummy.results", data: { winner: "p1", scores: [] } }, 0, "p1");
    expect(s.results?.winner).toBe("p1");
    s = rummyReducer(s, { type: "rummy.status", data: { playing: ["p2"], tables: 0 } }, 0, "p1");
    expect(s.match).toBeNull();
    expect(s.results).toBeNull();
  });

  it("closes the list when a match starts, and remembers rejections", () => {
    let s = setListOpen(initialRummyState, true);
    s = rummyReducer(s, { type: "rummy.rejected", data: { reason: "full" } }, 0, "p1");
    expect(s.rejection).toBe("full");
    s = rummyReducer(s, state(), 0, "p1");
    expect(s.listOpen).toBe(false);
    expect(s.rejection).toBeNull();
  });

  it("finds the room a player is in", () => {
    const s = rummyReducer(
      initialRummyState,
      {
        type: "rummy.rooms",
        data: { rooms: [{ id: "s2", mode: "free", players: ["p1", "p3"], phase: "lobby", lobbyMs: 0 }] },
      },
      0,
      "p1",
    );
    expect(roomOf(s, "p3")?.id).toBe("s2");
    expect(roomOf(s, "p9")).toBeNull();
    expect(s.rooms[0].lobbyEndsAt).toBeNull();
  });
});
