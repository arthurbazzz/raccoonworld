// What each tile ID (0-105) stands for: the same fixed table as the Go
// domain (server/internal/domain/rummy/tile.go, TileOf).

/** Colors in the order of the art rows (spec, Q6). */
export const COLORS = ["brown", "red", "green", "blue"] as const;
export type TileColor = (typeof COLORS)[number];

export const MAX_VALUE = 13;
export const DECK_SIZE = 2 * COLORS.length * MAX_VALUE + 2;
export const JOKER_PENALTY = 30;

export interface Tile {
  id: number;
  color: TileColor | null; // null for a joker
  value: number; // 1-13; 0 for a joker
  joker: boolean;
}

/** IDs are laid out as copy*52 + color*13 + (value-1); 104 and 105 are jokers. */
export function tileOf(id: number): Tile {
  if (id >= 2 * COLORS.length * MAX_VALUE) return { id, color: null, value: 0, joker: true };
  const n = id % (COLORS.length * MAX_VALUE);
  return { id, color: COLORS[Math.floor(n / MAX_VALUE)], value: (n % MAX_VALUE) + 1, joker: false };
}

/** "red-7" or "joker", as in the Go tests and the shared meld cases. */
export function faceOf(id: number): string {
  const t = tileOf(id);
  return t.joker ? "joker" : `${t.color}-${t.value}`;
}

const COLOR_NAMES: Record<TileColor, string> = {
  brown: "marrom",
  red: "vermelho",
  green: "verde",
  blue: "azul",
};

/** How a player reads the tile, for labels and screen readers: "7 vermelho". */
export function tileName(id: number): string {
  const t = tileOf(id);
  return t.joker ? "coringa" : `${t.value} ${COLOR_NAMES[t.color!]}`;
}

/** Hand order by color, then number (RN26). */
export function byColor(a: number, b: number): number {
  const ta = tileOf(a);
  const tb = tileOf(b);
  if (ta.joker !== tb.joker) return ta.joker ? 1 : -1;
  return (
    COLORS.indexOf(ta.color ?? "brown") - COLORS.indexOf(tb.color ?? "brown") ||
    ta.value - tb.value ||
    a - b
  );
}

/** Hand order by number, then color (RN26). */
export function byNumber(a: number, b: number): number {
  const ta = tileOf(a);
  const tb = tileOf(b);
  if (ta.joker !== tb.joker) return ta.joker ? 1 : -1;
  return (
    ta.value - tb.value ||
    COLORS.indexOf(ta.color ?? "brown") - COLORS.indexOf(tb.color ?? "brown") ||
    a - b
  );
}
