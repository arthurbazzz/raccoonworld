// Meld rules (RN2, RN3), the same as the Go domain (meld.go). The client
// uses them to show what is wrong while the player arranges the table
// (RN12, RN12a); the server decides.

import type { RummyIssue } from "@/net/protocol";

import { COLORS, MAX_VALUE, tileOf, type Tile } from "./tiles";

export type MeldKind = "group" | "run";
export type MeldError =
  "too_short" | "too_long" | "repeated_color" | "mixed_colors" | "gap" | "out_of_range";

export type Classified = { kind: MeldKind } | { error: MeldError };

/** What is wrong with one meld: the code, the tiles to blame and the numbers. */
export type Diagnosis = { kind: MeldKind } | { issue: Omit<RummyIssue, "meld"> };

const MIN_MELD = 3;
export const OPENING_POINTS = 30;

function firstReal(tiles: Tile[]): Tile {
  return tiles.find((t) => !t.joker) ?? tiles[0];
}

function runStart(tiles: Tile[]): number {
  const i = tiles.findIndex((t) => !t.joker);
  return i < 0 ? 0 : tiles[i].value - i;
}

type Issue = Omit<RummyIssue, "meld">;

const colorIndex = (t: Tile) => COLORS.indexOf(t.color!);

function checkRun(tiles: Tile[]): Issue | null {
  const first = firstReal(tiles);
  const off = tiles.filter((t) => !t.joker && t.color !== first.color).map((t) => t.id);
  if (off.length > 0)
    return { code: "mixed_colors", tiles: off, params: { color: colorIndex(first) } };
  const start = runStart(tiles);
  const i = tiles.findIndex((t, i) => !t.joker && t.value !== start + i);
  if (i >= 0) {
    return {
      code: "gap",
      tiles: [tiles[i].id],
      params: { color: colorIndex(first), expected: start + i, found: tiles[i].value },
    };
  }
  if (start < 1 || start + tiles.length - 1 > MAX_VALUE) {
    return { code: "out_of_range", tiles: [], params: { start, end: start + tiles.length - 1 } };
  }
  return null;
}

function checkGroup(tiles: Tile[]): Issue | null {
  if (tiles.length > COLORS.length) {
    return { code: "too_long", tiles: [], params: { count: tiles.length } };
  }
  const seen = new Set<string>();
  for (const t of tiles) {
    if (t.joker) continue;
    if (seen.has(t.color!)) {
      return {
        code: "repeated_color",
        tiles: tiles.filter((o) => !o.joker && o.color === t.color).map((o) => o.id),
        params: { color: colorIndex(t), value: t.value },
      };
    }
    seen.add(t.color!);
  }
  return null;
}

/**
 * Whether tiles, in their order on the table, form a meld, and if not, why
 * (RN12a). Tiles that fit both kinds are a run if their order allows, as on
 * the server.
 */
export function diagnose(ids: number[]): Diagnosis {
  if (ids.length < MIN_MELD) {
    return { issue: { code: "too_short", tiles: [], params: { count: ids.length } } };
  }
  const tiles = ids.map(tileOf);
  const runIssue = checkRun(tiles);
  if (runIssue === null) return { kind: "run" };
  const v = firstReal(tiles).value;
  if (tiles.some((t) => !t.joker && t.value !== v)) return { issue: runIssue };
  const groupIssue = checkGroup(tiles);
  return groupIssue === null ? { kind: "group" } : { issue: groupIssue };
}

/** Whether tiles form a meld (RN2), without the details. */
export function classify(ids: number[]): Classified {
  const d = diagnose(ids);
  return "kind" in d ? d : { error: d.issue.code as MeldError };
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

const COLOR_ONE = ["marrom", "vermelho", "verde", "azul"];
const COLOR_MANY = ["marrons", "vermelhos", "verdes", "azuis"];
const COLOR_FEM_MANY = ["marrons", "vermelhas", "verdes", "azuis"];

/** "7 verde", for a tile named by its numbers. */
const tileWords = (value: number, color: number) => `${value} ${COLOR_ONE[color]}`;

/**
 * The sentence a player reads for an issue (RN12a): what is wrong and how
 * to fix it. The server only sends codes; the words are here.
 */
export function describeIssue(is: Pick<RummyIssue, "code" | "tiles" | "params">): string {
  const p = is.params;
  switch (is.code) {
    case "too_short":
      return p.count === 1
        ? "Esta combinação tem só 1 peça; precisa de pelo menos 3."
        : `Esta combinação tem só ${p.count} peças; precisa de pelo menos 3.`;
    case "too_long":
      return "Um grupo tem no máximo 4 peças, uma de cada cor.";
    case "repeated_color":
      return `O grupo de ${p.value} tem dois ${p.value} ${COLOR_MANY[p.color]}. Tire um deles.`;
    case "mixed_colors": {
      const off = is.tiles.map(tileOf).find((t) => !t.joker);
      const which = off ? `o ${tileWords(off.value, colorIndex(off))}` : "uma peça";
      return `Sequência com cores diferentes: ${which} não combina com as ${COLOR_FEM_MANY[p.color]}.`;
    }
    case "gap":
      return p.expected > MAX_VALUE
        ? `Depois do 13 a sequência acaba; o ${p.found} não pode vir aí.`
        : `Depois do ${p.expected - 1} ${COLOR_ONE[p.color]} vem o ${p.expected}, não o ${p.found}. Falta o ${tileWords(p.expected, p.color)}.`;
    case "out_of_range":
      return p.start < 1
        ? "A sequência começa antes do 1. Tire o coringa do começo."
        : "A sequência passou do 13. Tire o coringa do fim.";
    case "opening_too_low":
      return `Sua abertura soma ${p.points} pontos; faltam ${p.need} para ${OPENING_POINTS}.`;
    case "opening_used_table": {
      const t = is.tiles.map(tileOf).find((t) => !t.joker);
      return t
        ? `Na abertura use só peças do seu suporte; o ${tileWords(t.value, colorIndex(t))} é da mesa.`
        : "Na abertura use só peças do seu suporte e não mexa nas combinações da mesa.";
    }
    case "nothing_placed":
      return "Você não baixou nenhuma peça. Baixe uma combinação ou clique em Comprar.";
    default:
      return "Jogada inválida.";
  }
}

/**
 * What the server would refuse in this draft (RN12a, RN13): every invalid
 * meld, or else what breaks the opening. `startMelds` and `startHand` are the
 * table and hand at the start of the turn.
 */
export function checkDraft(
  melds: number[][],
  startMelds: number[][],
  startHand: number[],
  opening: boolean,
): RummyIssue[] {
  const issues: RummyIssue[] = [];
  melds.forEach((meld, i) => {
    const d = diagnose(meld);
    if ("issue" in d) issues.push({ meld: i, ...d.issue });
  });
  if (issues.length > 0 || !opening) return issues;

  const left = [...startMelds];
  let total = 0;
  melds.forEach((meld, i) => {
    const j = left.findIndex((o) => o.length === meld.length && o.every((t, k) => t === meld[k]));
    if (j >= 0) {
      left.splice(j, 1);
      return;
    }
    const fromTable = meld.filter((t) => !startHand.includes(t));
    if (fromTable.length > 0) {
      issues.push({ meld: i, code: "opening_used_table", tiles: fromTable, params: {} });
    }
    total += points(meld);
  });
  if (issues.length === 0 && left.length > 0) {
    issues.push({ meld: -1, code: "opening_used_table", tiles: left.flat(), params: {} });
  }
  if (issues.length === 0 && total < OPENING_POINTS) {
    issues.push({
      meld: -1,
      code: "opening_too_low",
      tiles: [],
      params: { points: total, need: OPENING_POINTS - total },
    });
  }
  return issues;
}
