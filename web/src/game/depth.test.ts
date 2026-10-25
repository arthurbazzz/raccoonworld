import { describe, expect, it } from "vitest";

import { playerDepth, tileDepth } from "./depth";

// RN6: whoever is lower on screen is drawn in front.
describe("depth", () => {
  const tile = { col: 5, row: 5, x: 160, y: 160, size: 32 }; // bottom edge at y=192

  it("draws a raccoon standing above a tile behind it", () => {
    expect(playerDepth(160)).toBeLessThan(tileDepth(tile));
  });

  it("draws a raccoon standing flush below a tile in front of it", () => {
    expect(playerDepth(192)).toBeGreaterThan(tileDepth(tile));
  });

  it("orders raccoons by feet position", () => {
    expect(playerDepth(300)).toBeGreaterThan(playerDepth(299.5));
  });
});
