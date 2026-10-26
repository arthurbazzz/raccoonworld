import { describe, expect, it } from "vitest";

import type { ServerMessage } from "@/net/protocol";

import { INTERPOLATION_DELAY_MS, Interpolator, TICK_MS } from "./interpolation";

const D = INTERPOLATION_DELAY_MS;

function snapshot(tick: number, ...players: [string, number, number][]): ServerMessage {
  return {
    type: "room.snapshot",
    data: { tick, players: players.map(([id, x, y]) => ({ id, x, y })) },
  };
}

function joined(...players: [string, number, number][]): ServerMessage {
  return {
    type: "room.joined",
    data: {
      you: players[0][0],
      map: "praca",
      players: players.map(([id, x, y]) => ({ id, nickname: id, x, y })),
    },
  };
}

describe("Interpolator", () => {
  it("shows a player where they joined", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 100, 200]), 0);
    expect(it.position("a", 0)).toEqual({ x: 100, y: 200 });
    expect(it.position("a", 5000)).toEqual({ x: 100, y: 200 });
  });

  it("renders INTERPOLATION_DELAY_MS in the past, blending two snapshots", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 0, 0]), 0);
    it.ingest(snapshot(1, ["a", 8, 0]), 50); // moving from rest
    it.ingest(snapshot(2, ["a", 16, 0]), 100);

    // At now = 100 + D the render time is 100: exactly the last snapshot.
    expect(it.position("a", 100 + D)).toEqual({ x: 16, y: 0 });
    // Halfway between the snapshots received at 50 and 100.
    expect(it.position("a", 75 + D)).toEqual({ x: 12, y: 0 });
  });

  it("starts moving from rest instead of gliding from an old sample", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 0, 0]), 0);
    // Idle for 10s (no snapshots), then a first move arrives.
    it.ingest(snapshot(200, ["a", 8, 0]), 10_000);

    // One tick before the move the player is still at rest...
    expect(it.position("a", 10_000 - TICK_MS + D)).toEqual({ x: 0, y: 0 });
    // ...and arrives at the new position when the move is rendered.
    expect(it.position("a", 10_000 + D)).toEqual({ x: 8, y: 0 });
  });

  it("holds the last position instead of extrapolating", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 0, 0]), 0);
    it.ingest(snapshot(1, ["a", 8, 0]), 50);
    expect(it.position("a", 60_000)).toEqual({ x: 8, y: 0 });
  });

  it("tracks arrivals and departures", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 0, 0]), 0);
    it.ingest({ type: "player.joined", data: { id: "b", nickname: "b", x: 5, y: 6 } }, 10);
    expect(it.position("b", 10)).toEqual({ x: 5, y: 6 });

    it.ingest({ type: "player.left", data: { id: "b" } }, 20);
    expect(it.position("b", 20)).toBeNull();
  });

  it("forgets everyone on a new room.joined (e.g. after reconnecting)", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 0, 0], ["b", 1, 1]), 0);
    it.ingest(joined(["c", 2, 2]), 100);
    expect(it.position("a", 100)).toBeNull();
    expect(it.position("c", 100)).toEqual({ x: 2, y: 2 });
  });

  it("keeps a bounded history per player", () => {
    const it = new Interpolator();
    it.ingest(joined(["a", 0, 0]), 0);
    for (let i = 1; i <= 1000; i++) it.ingest(snapshot(i, ["a", i, 0]), i * TICK_MS);
    expect(it.historySize("a")).toBeLessThanOrEqual(32);
    expect(it.position("a", 1000 * TICK_MS + D)).toEqual({ x: 1000, y: 0 });
  });

  describe("motion (RW-002)", () => {
    // Walks right: 8px per tick, snapshots at 50 and 100.
    function walkingRight() {
      const it = new Interpolator();
      it.ingest(joined(["a", 0, 0]), 0);
      it.ingest(snapshot(1, ["a", 8, 0]), 50);
      it.ingest(snapshot(2, ["a", 16, 0]), 100);
      return it;
    }

    it("starts facing down and standing still (RV1)", () => {
      const it = new Interpolator();
      it.ingest(joined(["a", 0, 0]), 0);
      expect(it.motion("a", 5000)).toEqual({ facing: "down", moving: false });
    });

    it("faces the direction of the move while walking", () => {
      expect(walkingRight().motion("a", 100 + D)).toEqual({ facing: "right", moving: true });
    });

    it("keeps walking for 150ms after the last move, then stands still (RV2)", () => {
      const w = walkingRight();
      // Render time 249: the position at 99 was still changing.
      expect(w.motion("a", 249 + D)).toEqual({ facing: "right", moving: true });
      // Render time 250: unchanged since 100.
      expect(w.motion("a", 250 + D)).toEqual({ facing: "right", moving: false });
    });

    it("keeps facing where it walked after stopping (RV1)", () => {
      expect(walkingRight().motion("a", 60_000)).toEqual({ facing: "right", moving: false });
    });

    it("uses the latest direction when the move changes", () => {
      const it = new Interpolator();
      it.ingest(joined(["a", 0, 0]), 0);
      it.ingest(snapshot(1, ["a", 0, -8]), 50); // up
      it.ingest(snapshot(2, ["a", -8, -8]), 100); // then left
      expect(it.motion("a", 100 + D)?.facing).toBe("left");
      // Halfway through the first segment it was still facing up.
      expect(it.motion("a", 25 + D)?.facing).toBe("up");
    });

    it("uses the side on a diagonal", () => {
      const it = new Interpolator();
      it.ingest(joined(["a", 0, 0]), 0);
      it.ingest(snapshot(1, ["a", 5.66, -5.66]), 50);
      expect(it.motion("a", 50 + D)?.facing).toBe("right");
    });

    it("does not turn before the move is rendered", () => {
      const it = new Interpolator();
      it.ingest(joined(["a", 0, 0]), 0);
      it.ingest(snapshot(1, ["a", 8, 0]), 10_000); // rest sample at 9950
      expect(it.motion("a", 9949 + D)).toEqual({ facing: "down", moving: false });
    });

    it("returns null for an unknown player", () => {
      expect(new Interpolator().motion("ghost", 0)).toBeNull();
    });
  });

  it("uses a delay of about two ticks", () => {
    expect(D).toBe(100);
    expect(TICK_MS).toBe(50);
  });
});
