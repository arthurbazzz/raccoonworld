import type { PlayerId } from "@/net/protocol";

import type { Interpolator } from "./interpolation";
import type { Motion } from "./motion";

/**
 * How often the accessible player list picks up facing and walking state
 * (plan RW-002, decision 2): often enough for tests and screen readers,
 * rare enough not to re-render React every frame.
 */
export const MOTION_REFRESH_MS = 100;

export type Motions = Record<PlayerId, Motion>;

/** The motion of each listed player at `now`, from the same source the scene draws. */
export function readMotions(interpolator: Interpolator, ids: PlayerId[], now: number): Motions {
  const out: Motions = {};
  for (const id of ids) {
    const m = interpolator.motion(id, now);
    if (m) out[id] = m;
  }
  return out;
}

export function sameMotions(a: Motions, b: Motions): boolean {
  const ids = Object.keys(a);
  if (ids.length !== Object.keys(b).length) return false;
  return ids.every(
    (id) => b[id] !== undefined && a[id].facing === b[id].facing && a[id].moving === b[id].moving,
  );
}
