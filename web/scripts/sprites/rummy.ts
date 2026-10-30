// Finds the pieces of the AI-made Rummy das Nozes sheet (RW-006, Q6).
// Pure logic: rummy-build.ts does the I/O.
//
// The sheet is 5 rows of 3 frames. Rows 1-4, one per color (brown, red,
// green, blue): the back, a stack of backs and a strip of 14 touching tiles
// (1 to 13 and the joker). Row 5: the table, a big back and a big stack.
// The strip is cut where the tile outlines meet: dark columns with fewer
// opaque pixels than the inside of a tile (the rounded corners).
//
// Runs with Node's built-in TypeScript: no enums, no constructor parameter
// properties.

import { findFrames, SpriteSheetError, type FrameBox, type RGBAImage } from "./slice.ts";

/** A piece's box in the source sheet, in pixels. */
export interface Box {
  x: number;
  y: number;
  w: number;
  h: number;
}

/** Where each piece is in the sheet (rummy.json). */
export interface RummyManifest {
  image: string;
  /** tiles[color][value - 1], colors in the order brown, red, green, blue. */
  tiles: Box[][];
  joker: Box;
  back: Box;
  stack: Box;
  table: Box;
}

const COLORS = 4;
const TILES_PER_STRIP = 14; // 13 numbers and the joker
const ALPHA = 128;
/** A column is a cut candidate below this share of the strip's fullest column... */
const VALLEY = 0.93;
/** ...and when it is outline: this share of its opaque pixels is dark. */
const OUTLINE = 0.9;
const DARK = 250; // r + g + b below this is outline

const boxOf = (f: FrameBox): Box => ({
  x: f.x0,
  y: f.top,
  w: f.x1 - f.x0 + 1,
  h: f.bottom - f.top + 1,
});

/** Cuts a strip of touching tiles at the valleys of opaque pixels per column. */
export function splitStrip(img: RGBAImage, strip: FrameBox, expected: number): FrameBox[] {
  const { width, data } = img;
  const counts: number[] = [];
  const outline: boolean[] = [];
  for (let x = strip.x0; x <= strip.x1; x++) {
    let n = 0;
    let dark = 0;
    for (let y = strip.top; y <= strip.bottom; y++) {
      const i = (y * width + x) * 4;
      if (data[i + 3] < ALPHA) continue;
      n++;
      if (data[i] + data[i + 1] + data[i + 2] < DARK) dark++;
    }
    counts.push(n);
    outline.push(n > 0 && dark >= n * OUTLINE);
  }
  const limit = Math.max(...counts) * VALLEY;

  // Runs of low columns away from the strip edges; cut at each run's lowest.
  const cuts: number[] = [];
  let run: number[] = [];
  const flush = (atEdge: boolean) => {
    if (run.length > 0 && !atEdge) {
      cuts.push(run.reduce((best, i) => (counts[i] < counts[best] ? i : best), run[0]));
    }
    run = [];
  };
  let leading = true;
  counts.forEach((n, i) => {
    if (n < limit && outline[i]) {
      run.push(i);
    } else {
      flush(leading);
      leading = false;
    }
  });
  run = []; // a run touching the right edge is the last tile's outline

  const edges = [0, ...cuts, counts.length];
  if (edges.length - 1 !== expected) {
    throw new SpriteSheetError(`esperado ${expected} peças na fileira, achei ${edges.length - 1}`);
  }
  return edges.slice(0, -1).map((start, i) => ({
    x0: strip.x0 + start,
    x1: strip.x0 + edges[i + 1] - 1,
    top: strip.top,
    bottom: strip.bottom,
  }));
}

export function rummyManifest(img: RGBAImage, image: string): RummyManifest {
  const frames = findFrames(img, { rows: COLORS + 1, cols: 3 });
  const tiles: Box[][] = [];
  let joker: Box | null = null;
  for (let c = 0; c < COLORS; c++) {
    const pieces = splitStrip(img, frames[c][2], TILES_PER_STRIP).map(boxOf);
    tiles.push(pieces.slice(0, 13));
    joker ??= pieces[13]; // every row has the same joker; the first is used
  }
  return {
    image,
    tiles,
    joker: joker!,
    // Only the brown back is used: a back per color would show the color of
    // hidden tiles (spec, Q6).
    back: boxOf(frames[0][0]),
    stack: boxOf(frames[0][1]),
    table: boxOf(frames[COLORS][0]),
  };
}
