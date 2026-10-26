// The description of the aligned raccoon sprite sheet, written by
// `npm run sprites` (web/scripts/sprites) and read by the game (RW-004).

export interface SheetManifest {
  /** URL of the aligned sheet, served from web/public. */
  image: string;
  /** Size of every cell; the sheet is a grid of cells. */
  cell: { w: number; h: number };
  /** Where the feet are inside a cell: bottom center, on the line under the feet. */
  anchor: { x: number; y: number };
  /** Which row holds each pose. */
  rows: { front: number; back: number; side: number };
  /** Columns of the idle frames. */
  idle: number[];
  /** Columns of the walk cycle, in order. */
  walk: number[];
  /** Typical height of the art in source pixels, used to scale it in the room. */
  artHeight: number;
}
