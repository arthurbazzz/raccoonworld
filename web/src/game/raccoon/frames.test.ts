import { describe, expect, it } from "vitest";

import { frameFor, IDLE_FRAME_MS, WALK_FPS } from "./frames";
import type { SheetManifest } from "./manifest";

const manifest: SheetManifest = {
  image: "/sprites/raccoon.png",
  cell: { w: 109, h: 120 },
  anchor: { x: 54, y: 116 },
  rows: { front: 0, back: 1, side: 2 },
  idle: [0, 1],
  walk: [2, 3, 4, 5],
  artHeight: 108,
};

describe("frameFor (RS5)", () => {
  it.each([
    ["down", 0, false],
    ["up", 1, false],
    ["right", 2, false],
    ["left", 2, true],
  ] as const)("facing %s uses row %d (mirrored: %s)", (facing, row, mirrored) => {
    expect(frameFor(manifest, facing, false, 0)).toMatchObject({ row, mirrored });
  });

  it("walks through the 4 walk frames at 12 frames per second", () => {
    expect(WALK_FPS).toBe(12);
    const step = 1000 / WALK_FPS;
    const cols = [0, 1, 2, 3, 4, 5].map((i) => frameFor(manifest, "down", true, i * step + 1).col);
    expect(cols).toEqual([2, 3, 4, 5, 2, 3]);
  });

  it("alternates the 2 idle frames every second, like breathing", () => {
    expect(IDLE_FRAME_MS).toBe(1000);
    const cols = [0, 999, 1000, 1999, 2000].map((t) => frameFor(manifest, "up", false, t).col);
    expect(cols).toEqual([0, 0, 1, 1, 0]);
  });

  it("follows the manifest, so new art with other columns works without code changes (RS6)", () => {
    const other = { ...manifest, idle: [5], walk: [0, 1] };
    expect(frameFor(other, "down", false, 12_345).col).toBe(5);
    expect(frameFor(other, "down", true, 0).col).toBe(0);
  });
});
