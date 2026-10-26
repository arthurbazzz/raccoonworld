// Which way a raccoon faces and whether it is walking (RW-002), derived only
// from positions so every client sees the same thing (RV5).

/** The four poses of the 3/4 view: down = front, up = back. */
export type Facing = "down" | "up" | "left" | "right";

export interface Motion {
  facing: Facing;
  moving: boolean;
}

/** Steps smaller than this (in room pixels) are floating point noise. */
export const MIN_STEP_PX = 0.01;

/**
 * RV1: any horizontal movement shows the side (so diagonals use the side,
 * the view that reads best on screen); otherwise front or back. A step
 * with no real movement keeps the previous facing.
 */
export function facingFrom(dx: number, dy: number, previous: Facing): Facing {
  if (Math.abs(dx) >= MIN_STEP_PX) return dx > 0 ? "right" : "left";
  if (Math.abs(dy) >= MIN_STEP_PX) return dy > 0 ? "down" : "up";
  return previous;
}
