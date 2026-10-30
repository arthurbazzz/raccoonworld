import { describe, expect, it } from "vitest";

import type { CupsRoundData, ServerMessage } from "@/net/protocol";

import { cupsReducer, cupsRole, initialCupsState, pickLocally } from "./cupsState";

const round: CupsRoundData = {
  round: 1,
  total: 5,
  nut: 0,
  showMs: 1500,
  swaps: [[0, 1]],
  swapMs: 600,
  pickMs: 10_000,
  elapsedMs: 0,
};

const status = (
  phase: "idle" | "lobby" | "playing" | "results",
  players: string[] = [],
  queued: string[] = [],
  lobbyMs = 0,
): ServerMessage => ({ type: "cups.status", data: { phase, players, queued, lobbyMs } });

const apply = (msgs: ServerMessage[], at = 1000) =>
  msgs.reduce((s, m) => cupsReducer(s, m, at), initialCupsState);

describe("cupsReducer", () => {
  it("counts the lobby down from when the status arrived (RC6)", () => {
    const s = apply([status("lobby", ["c1", "c2"], [], 6000)], 500);
    expect(s).toMatchObject({ phase: "lobby", players: ["c1", "c2"], lobbyEndsAt: 6500 });
  });

  it("starts a round at its arrival, and a joined snapshot back-dates it", () => {
    expect(apply([{ type: "cups.round", data: round }], 2000).round?.startedAt).toBe(2000);
    const joined: ServerMessage = {
      type: "room.joined",
      data: {
        you: "c3",
        map: "praca",
        players: [],
        cups: {
          phase: "playing",
          players: ["c1"],
          queued: [],
          lobbyMs: 0,
          round: { ...round, elapsedMs: 700 },
        },
      },
    };
    const s = apply([joined], 2000);
    expect(s.phase).toBe("playing");
    expect(s.round?.startedAt).toBe(1300);
  });

  it("keeps who picked without the cup, and our own pick only once", () => {
    let s = apply([status("playing", ["c1", "c2"]), { type: "cups.round", data: round }]);
    s = cupsReducer(s, { type: "cups.picked", data: { id: "c2" } }, 0);
    s = pickLocally(pickLocally(s, 2), 0);
    expect(s.picked).toEqual(["c2"]);
    expect(s.myPick).toBe(2);
  });

  it("clears picks and the reveal on a new round", () => {
    let s = apply([
      { type: "cups.round", data: round },
      { type: "cups.picked", data: { id: "c1" } },
      { type: "cups.reveal", data: { round: 1, nut: 1, picks: [] } },
    ]);
    s = pickLocally(s, 1);
    s = cupsReducer(s, { type: "cups.round", data: { ...round, round: 2 } }, 0);
    expect(s).toMatchObject({ picked: [], myPick: null, reveal: null });
  });

  it("keeps the scoreboard while results are up and forgets it on the next lobby", () => {
    let s = apply([
      status("playing", ["c1"]),
      { type: "cups.results", data: { scores: [{ id: "c1", hits: 4, place: 1 }] } },
      status("results", ["c1"]),
    ]);
    expect(s.scores).toHaveLength(1);
    s = cupsReducer(s, status("idle"), 0);
    expect(s).toMatchObject({ phase: "idle", scores: null, round: null });
  });

  it("remembers a refused E until the next status", () => {
    let s = apply([{ type: "cups.rejected", data: { reason: "too_far" } }]);
    expect(s.tooFar).toBe(true);
    s = cupsReducer(s, status("lobby", ["c1"], [], 10_000), 0);
    expect(s.tooFar).toBe(false);
  });

  it("ignores messages about the room", () => {
    const s = apply([status("lobby", ["c1"])]);
    expect(cupsReducer(s, { type: "player.left", data: { id: "c9" } }, 0)).toBe(s);
  });
});

describe("cupsRole", () => {
  it("tells players, queued and others apart (RC5, RC22)", () => {
    const s = apply([status("playing", ["c1"], ["c2"])]);
    expect(cupsRole(s, "c1")).toBe("player");
    expect(cupsRole(s, "c2")).toBe("queued");
    expect(cupsRole(s, "c3")).toBeNull();
    expect(cupsRole(s, null)).toBeNull();
  });
});
