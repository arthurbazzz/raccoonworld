import { describe, expect, it } from "vitest";

import { BOOTH_REACH, nearBooth } from "./booth";

// Same cases as TestNearBooth in the Go domain: the box (64,32)-(128,64).
const booth = { x: 64, y: 32, w: 64, h: 32 };
const R = BOOTH_REACH;

describe("nearBooth (RC2)", () => {
  it.each([
    ["below, touching", 96, 64, true],
    ["below, at reach", 96, 64 + R, true],
    ["below, past reach", 96, 64 + R + 1, false],
    ["above, at reach", 96, 32 - R, true],
    ["left, at reach", 64 - R, 48, true],
    ["left, past reach", 64 - R - 1, 48, false],
    ["right, at reach", 128 + R, 48, true],
    ["right, past reach", 128 + R + 1, 48, false],
    ["corner, inside the circle", 128 + 30, 64 + 30, true],
    ["corner, outside the circle", 128 + 35, 64 + 35, false],
  ] as const)("%s", (_, x, y, want) => {
    expect(nearBooth(booth, x, y)).toBe(want);
  });

  it("is false when the map has no booth", () => {
    expect(nearBooth(null, 96, 64)).toBe(false);
  });

  it("uses the server reach of 48 px", () => {
    expect(BOOTH_REACH).toBe(48);
  });
});
