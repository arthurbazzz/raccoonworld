// Lays the detected frames out in a regular grid with the feet of every
// frame on the same point (RW-004, RS3). Pure logic: build.ts does the I/O.

import type { SheetManifest } from "../../src/game/raccoon/manifest.ts";
import type { FrameBox } from "./slice.ts";

/** Empty space around the biggest frame in each cell. */
export const PAD = 4;

/** Where one source frame goes in the aligned sheet. */
export interface Placement {
  row: number;
  col: number;
  src: FrameBox;
  /** Top-left corner in the aligned sheet. */
  x: number;
  y: number;
}

export interface SheetLayout {
  cell: { w: number; h: number };
  anchor: { x: number; y: number };
  size: { w: number; h: number };
  placements: Placement[];
  artHeight: number;
}

const widthOf = (f: FrameBox) => f.x1 - f.x0 + 1;
const heightOf = (f: FrameBox) => f.bottom - f.top + 1;

function median(values: number[]): number {
  const s = [...values].sort((a, b) => a - b);
  const mid = s.length >> 1;
  return s.length % 2 ? s[mid] : (s[mid - 1] + s[mid]) / 2;
}

export function layoutSheet(frames: FrameBox[][]): SheetLayout {
  const all = frames.flat();
  const cell = {
    w: Math.max(...all.map(widthOf)) + 2 * PAD,
    h: Math.max(...all.map(heightOf)) + 2 * PAD,
  };
  // The anchor is the line right under the feet, centered.
  const anchor = { x: Math.floor(cell.w / 2), y: cell.h - PAD };

  const placements = frames.flatMap((row, r) =>
    row.map((src, c) => ({
      row: r,
      col: c,
      src,
      x: c * cell.w + anchor.x - Math.floor(widthOf(src) / 2),
      y: r * cell.h + anchor.y - heightOf(src),
    })),
  );

  const cols = Math.max(...frames.map((r) => r.length));
  return {
    cell,
    anchor,
    size: { w: cols * cell.w, h: frames.length * cell.h },
    placements,
    artHeight: median(all.map(heightOf)),
  };
}

/** The manifest the game reads, for the agreed sheet layout (spec RW-004). */
export function manifestFor(layout: SheetLayout, image: string): SheetManifest {
  return {
    image,
    cell: layout.cell,
    anchor: layout.anchor,
    rows: { front: 0, back: 1, side: 2 },
    idle: [0, 1],
    walk: [2, 3, 4, 5],
    artHeight: layout.artHeight,
  };
}
