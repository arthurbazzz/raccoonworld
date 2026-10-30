import { describe, expect, it } from "vitest";

import { nearTable, TABLE_REACH } from "./table";

// Same cases as TestNearTable in the Go domain: the box (32,64)-(96,128).
const table = { x: 32, y: 64, w: 64, h: 64 };
const R = TABLE_REACH;

describe("nearTable (RN4)", () => {
  it.each([
    ["below, at reach", 64, 128 + R, true],
    ["below, past reach", 64, 128 + R + 1, false],
    ["right, at reach", 96 + R, 96, true],
    ["right, past reach", 96 + R + 1, 96, false],
    ["corner, outside the circle", 96 + 35, 128 + 35, false],
  ] as const)("%s", (_, x, y, want) => {
    expect(nearTable(table, x, y)).toBe(want);
  });

  it("is false when the map has no table", () => {
    expect(nearTable(null, 64, 128)).toBe(false);
  });
});
