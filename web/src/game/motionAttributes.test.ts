import { describe, expect, it } from "vitest";

import { Interpolator } from "./interpolation";
import { MOTION_REFRESH_MS, readMotions, sameMotions } from "./motionAttributes";

function room() {
  const it = new Interpolator();
  it.ingest(
    {
      type: "room.joined",
      data: {
        you: "a",
        map: "praca",
        players: [
          { id: "a", nickname: "a", x: 0, y: 0 },
          { id: "b", nickname: "b", x: 50, y: 50 },
        ],
      },
    },
    0,
  );
  return it;
}

describe("readMotions", () => {
  it("reads every listed player's motion at `now`", () => {
    const it = room();
    it.ingest({ type: "room.snapshot", data: { tick: 1, players: [{ id: "a", x: 8, y: 0 }] } }, 50);

    expect(readMotions(it, ["a", "b"], 150)).toEqual({
      a: { facing: "right", moving: true },
      b: { facing: "down", moving: false },
    });
  });

  it("skips players the interpolator does not know", () => {
    expect(readMotions(room(), ["a", "ghost"], 0)).toEqual({
      a: { facing: "down", moving: false },
    });
  });
});

describe("sameMotions", () => {
  const still = { facing: "down", moving: false } as const;

  it("is true for equal snapshots, so React does not re-render", () => {
    expect(sameMotions({ a: still }, { a: { ...still } })).toBe(true);
  });

  it("is false when a player turns, starts walking, joins or leaves", () => {
    expect(sameMotions({ a: still }, { a: { ...still, facing: "left" } })).toBe(false);
    expect(sameMotions({ a: still }, { a: { ...still, moving: true } })).toBe(false);
    expect(sameMotions({ a: still }, { a: still, b: still })).toBe(false);
    expect(sameMotions({ a: still, b: still }, { a: still })).toBe(false);
  });
});

it("refreshes 10 times per second (plan, decision 2)", () => {
  expect(MOTION_REFRESH_MS).toBe(100);
});
