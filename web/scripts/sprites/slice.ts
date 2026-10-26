// Finds the frames of an AI-made sprite sheet (RW-004, RS2). The art does
// not follow a strict grid, so frames are found by the fully transparent
// gaps between them instead of by fixed cell sizes.
//
// Runs with Node's built-in TypeScript: no enums, no constructor parameter
// properties.

/** Raw pixels, 4 bytes per pixel (RGBA), row by row. */
export interface RGBAImage {
  width: number;
  height: number;
  data: Uint8Array;
}

/** A frame's tight bounding box, inclusive, in source pixels. */
export interface FrameBox {
  x0: number;
  x1: number;
  top: number;
  bottom: number;
}

export interface FindOptions {
  rows: number;
  cols: number;
  /** A pixel counts as part of a frame from this alpha up. */
  alpha?: number;
  /** Bands thinner than this are noise, not frames. */
  minSize?: number;
}

/** The sheet does not have the expected layout. */
export class SpriteSheetError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "SpriteSheetError";
  }
}

const DEFAULT_ALPHA = 128;
const DEFAULT_MIN_SIZE = 8;

/** Index ranges [start, end] where `filled(i)` is true, ignoring short runs. */
function bands(
  length: number,
  filled: (i: number) => boolean,
  minSize: number,
): [number, number][] {
  const out: [number, number][] = [];
  let start = -1;
  for (let i = 0; i <= length; i++) {
    const on = i < length && filled(i);
    if (on && start < 0) start = i;
    if (!on && start >= 0) {
      if (i - start >= minSize) out.push([start, i - 1]);
      start = -1;
    }
  }
  return out;
}

/**
 * Returns frames[row][col], top to bottom and left to right. Throws
 * SpriteSheetError, with what it found, unless the sheet has exactly
 * `rows` x `cols` frames.
 */
export function findFrames(img: RGBAImage, opts: FindOptions): FrameBox[][] {
  const threshold = opts.alpha ?? DEFAULT_ALPHA;
  const minSize = opts.minSize ?? DEFAULT_MIN_SIZE;
  const { width, height, data } = img;
  const solid = (x: number, y: number) => data[(y * width + x) * 4 + 3] >= threshold;

  const rowBands = bands(
    height,
    (y) => {
      for (let x = 0; x < width; x++) if (solid(x, y)) return true;
      return false;
    },
    minSize,
  );

  const frames = rowBands.map(([y0, y1]) =>
    bands(
      width,
      (x) => {
        for (let y = y0; y <= y1; y++) if (solid(x, y)) return true;
        return false;
      },
      minSize,
    ).map(([x0, x1]): FrameBox => {
      let top = y1;
      let bottom = y0;
      for (let y = y0; y <= y1; y++) {
        for (let x = x0; x <= x1; x++) {
          if (solid(x, y)) {
            if (y < top) top = y;
            if (y > bottom) bottom = y;
          }
        }
      }
      return { x0, x1, top, bottom };
    }),
  );

  const ok = frames.length === opts.rows && frames.every((r) => r.length === opts.cols);
  if (!ok) {
    const found =
      frames.length === 0
        ? "achei 0 fileiras"
        : `achei ${frames.length} fileiras com ${frames.map((r) => r.length).join(", ")} colunas`;
    throw new SpriteSheetError(`esperado ${opts.rows}×${opts.cols} quadros, ${found}`);
  }
  return frames;
}
