import type { SolidTile } from "./tilemap";

// 3/4 view depth (RN6): sprites are sorted by the y of their base, so
// whoever is lower on screen is drawn in front.

/** A solid tile's base is its bottom edge. */
export function tileDepth(tile: SolidTile): number {
  return tile.y + tile.size;
}

/**
 * A raccoon's base is its feet. The half-pixel bias puts a raccoon standing
 * flush against a tile's bottom edge in front of that tile.
 */
export function playerDepth(feetY: number): number {
  return feetY + 0.5;
}
