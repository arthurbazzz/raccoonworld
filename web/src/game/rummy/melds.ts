// Meld rules (RN2, RN3), the same as the Go domain (meld.go). The client only
// uses them to mark invalid melds while the player arranges the table
// (RN12); the server decides.

import { COLORS, MAX_VALUE, tileOf, type Tile } from "./tiles";

export type MeldKind = "group" | "run";
export type MeldError =
  "too_short" | "too_long" | "repeated_color" | "mixed_colors" | "gap" | "out_of_range";

export type Classified = { kind: MeldKind } | { error: MeldError };

const MIN_MELD = 3;

function firstReal(tiles: Tile[]): Tile {
  return tiles.find((t) => !t.joker) ?? tiles[0];
}

function runStart(tiles: Tile[]): number {
  const i = tiles.findIndex((t) => !t.joker);
  return i < 0 ? 0 : tiles[i].value - i;
}

function checkRun(tiles: Tile[]): MeldError | null {
  const color = firstReal(tiles).color;
  if (tiles.some((t) => !t.joker && t.color !== color)) return "mixed_colors";
  const start = runStart(tiles);
  if (tiles.some((t, i) => !t.joker && t.value !== start + i)) return "gap";
  if (start < 1 || start + tiles.length - 1 > MAX_VALUE) return "out_of_range";
  return null;
}

function checkGroup(tiles: Tile[]): MeldError | null {
  if (tiles.length > COLORS.length) return "too_long";
  const seen = new Set<string>();
  for (const t of tiles) {
    if (t.joker) continue;
    if (seen.has(t.color!)) return "repeated_color";
    seen.add(t.color!);
  }
  return null;
}

/**
 * Whether tiles, in their order on the table, form a meld. Tiles that fit
 * both kinds are a run if their order allows, as on the server.
 */
export function classify(ids: number[]): Classified {
  if (ids.length < MIN_MELD) return { error: "too_short" };
  const tiles = ids.map(tileOf);
  const runError = checkRun(tiles);
  if (runError === null) return { kind: "run" };
  const v = firstReal(tiles).value;
  if (tiles.some((t) => !t.joker && t.value !== v)) return { error: runError };
  const groupError = checkGroup(tiles);
  return groupError === null ? { kind: "group" } : { error: groupError };
}

/** Sum of a valid meld's numbers, jokers counting as what they stand for. */
export function points(ids: number[]): number {
  const c = classify(ids);
  if (!("kind" in c)) return 0;
  const tiles = ids.map(tileOf);
  if (c.kind === "group") return tiles.length * firstReal(tiles).value;
  const n = tiles.length;
  return n * runStart(tiles) + (n * (n - 1)) / 2;
}

/** What a player reads when a meld is invalid, or a play is refused. */
export const REASONS: Record<string, string> = {
  too_short: "Uma combinação precisa de pelo menos 3 peças.",
  too_long: "Um grupo tem no máximo 4 peças.",
  repeated_color: "Um grupo não pode repetir cor.",
  mixed_colors: "Uma sequência precisa ser toda da mesma cor.",
  gap: "Os números da sequência precisam ser seguidos.",
  out_of_range: "A sequência vai de 1 a 13.",
  nothing_placed: "Você não baixou nenhuma peça.",
  opening_too_low: "A primeira jogada precisa somar 30 pontos.",
  opening_used_table: "Na primeira jogada, use só peças da sua mão.",
};
