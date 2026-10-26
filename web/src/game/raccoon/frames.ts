import type { Facing } from "../motion";

import type { SheetManifest } from "./manifest";

/** Walk animation speed (RW-004, Q3): ~2 frames per step at 6 steps/s. */
export const WALK_FPS = 12;
/** How long each idle frame stays, alternating like a slow breath. */
export const IDLE_FRAME_MS = 1000;

export interface FrameRef {
  row: number;
  col: number;
  /** Left is the right-facing side row, mirrored (RW-002). */
  mirrored: boolean;
}

/** Which frame of the sheet to show for a pose at time `t` (ms), RW-004 RS5. */
export function frameFor(m: SheetManifest, facing: Facing, moving: boolean, t: number): FrameRef {
  const row = facing === "down" ? m.rows.front : facing === "up" ? m.rows.back : m.rows.side;
  const cycle = moving ? m.walk : m.idle;
  const index = moving ? Math.floor((t * WALK_FPS) / 1000) : Math.floor(t / IDLE_FRAME_MS);
  return { row, col: cycle[index % cycle.length], mirrored: facing === "left" };
}
