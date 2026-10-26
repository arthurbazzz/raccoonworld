import type { Rect } from "../tilemap";

/** How close the feet must be to the booth to play, as on the server (RC2). */
export const BOOTH_REACH = 48;

/**
 * Whether feet at (x, y) are within BOOTH_REACH of the booth: the distance
 * from the point to the nearest edge of the box (same rule as the Go
 * `TileMap.NearBooth`).
 */
export function nearBooth(booth: Rect | null, x: number, y: number): boolean {
  if (!booth) return false;
  const dx = Math.max(0, booth.x - x, x - (booth.x + booth.w));
  const dy = Math.max(0, booth.y - y, y - (booth.y + booth.h));
  return Math.hypot(dx, dy) <= BOOTH_REACH;
}
