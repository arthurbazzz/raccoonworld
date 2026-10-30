// The Rummy das Nozes art (RW-006, Q6), cut by `npm run sprites` into
// web/public/sprites/rummy.png and rummy.json.

/** A piece's box in the sheet, in pixels. */
export interface Box {
  x: number;
  y: number;
  w: number;
  h: number;
}

/** Mirrors scripts/sprites/rummy.ts (RummyManifest). */
export interface RummyManifest {
  image: string;
  size: { w: number; h: number };
  /** tiles[color][value - 1], colors in the order of COLORS. */
  tiles: Box[][];
  joker: Box;
  back: Box;
  stack: Box;
  table: Box;
}

export const RUMMY_MANIFEST_URL = "/sprites/rummy.json";

let loading: Promise<RummyManifest | null> | null = null;

/** Loads the manifest once. Resolves to null when the art is missing, so
 * the game still works with plain tiles. */
export function loadRummyManifest(url: string = RUMMY_MANIFEST_URL): Promise<RummyManifest | null> {
  loading ??= fetch(url)
    .then((res) => (res.ok ? (res.json() as Promise<RummyManifest>) : null))
    .catch(() => null);
  return loading;
}
