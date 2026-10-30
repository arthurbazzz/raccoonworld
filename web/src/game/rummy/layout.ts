// Our turn's arrangement of the table (RN11, RN32-RN36): pure, so the drag
// and drop in the scene only has to say what was dropped where. The server
// only ever receives the resulting layout.

import type { MeldPos } from "@/net/protocol";

export const TABLE_COLS = 48;
export const TABLE_ROWS = 28;
/** A tile is 2 cells wide and 3 tall; a meld has a 1-cell handle on its left. */
export const TILE_COLS = 2;
export const MELD_ROWS = 3;

export interface Layout {
  melds: number[][];
  pos: MeldPos[];
  hand: number[];
}

/** Where tiles are dropped: into a meld (before a tile, or at its end), a new meld at a cell, or the hand. */
export type DropTarget =
  | { kind: "meld"; index: number; before: number | null }
  | { kind: "new"; at: MeldPos }
  | { kind: "hand" };

/** Cells a meld of n tiles takes, handle included. */
export const meldCols = (n: number) => 1 + TILE_COLS * n;

interface Rect {
  x: number;
  y: number;
  w: number;
  h: number;
}

const rectOf = (meld: number[], [x, y]: MeldPos): Rect => ({
  x,
  y,
  w: meldCols(meld.length),
  h: MELD_ROWS,
});

/** Whether two rects overlap or touch (melds keep a 1-cell gap). */
const touch = (a: Rect, b: Rect) =>
  a.x < b.x + b.w + 1 && b.x < a.x + a.w + 1 && a.y < b.y + b.h && b.y < a.y + a.h;

const clampPos = (meld: number[], [x, y]: MeldPos): MeldPos => [
  Math.max(0, Math.min(TABLE_COLS - Math.min(meldCols(meld.length), TABLE_COLS), Math.round(x))),
  Math.max(0, Math.min(TABLE_ROWS - MELD_ROWS, Math.round(y))),
];

/**
 * The free cell nearest to `want` for melds[index], so melds never overlap
 * (RN32). If the table is full, `want` itself (clamped) is used.
 */
export function freeSpot(melds: number[][], pos: MeldPos[], index: number, want: MeldPos): MeldPos {
  const meld = melds[index];
  const start = clampPos(meld, want);
  const others = melds
    .map((m, i) => (i === index ? null : rectOf(m, pos[i])))
    .filter((r): r is Rect => r !== null);
  const fits = (p: MeldPos) => {
    const c = clampPos(meld, p);
    if (c[0] !== p[0] || c[1] !== p[1]) return false;
    const r = rectOf(meld, p);
    return !others.some((o) => touch(r, o));
  };
  // Rings of growing distance around the wanted cell.
  for (let d = 0; d < TABLE_COLS + TABLE_ROWS; d++) {
    let best: MeldPos | null = null;
    let bestDist = Infinity;
    for (let dy = -d; dy <= d; dy++) {
      for (let dx = -d; dx <= d; dx++) {
        if (Math.max(Math.abs(dx), Math.abs(dy)) !== d) continue;
        const p: MeldPos = [start[0] + dx, start[1] + dy];
        const dist = dx * dx + dy * dy;
        if (dist < bestDist && fits(p)) {
          best = p;
          bestDist = dist;
        }
      }
    }
    if (best) return best;
  }
  return start;
}

/** Places melds[index] at the free cell nearest `want`. */
function settle(l: Layout, index: number, want: MeldPos): Layout {
  const pos = [...l.pos];
  pos[index] = freeSpot(l.melds, pos, index, want);
  return { ...l, pos };
}

/** Drops melds left empty, with their positions. */
function prune(l: Layout): Layout {
  const keep = l.melds.map((m) => m.length > 0);
  return {
    melds: l.melds.filter((_, i) => keep[i]),
    pos: l.pos.filter((_, i) => keep[i]),
    hand: l.hand,
  };
}

/**
 * Moves tiles `ids`, in this order, to `target` (RN33, RN35). Melds left
 * empty disappear. A tile that was on the table at the start of the turn
 * cannot go back to the hand: `startHand` says which tiles may (RN11; the
 * server would refuse it anyway). Returns null when nothing changes or the
 * move is not allowed.
 */
export function moveTiles(
  layout: Layout,
  ids: number[],
  target: DropTarget,
  startHand: number[],
): Layout | null {
  if (ids.length === 0) return null;
  if (target.kind === "hand" && ids.some((id) => !startHand.includes(id))) return null;
  if (target.kind === "meld" && target.before !== null && ids.includes(target.before)) return null;
  const known = (id: number) =>
    layout.hand.includes(id) || layout.melds.some((m) => m.includes(id));
  if (!ids.every(known)) return null;
  if (target.kind === "hand" && ids.every((id) => layout.hand.includes(id))) return null;

  // Keep the target meld by identity: removing tiles may empty other melds.
  const targetMeld = target.kind === "meld" ? layout.melds[target.index] : null;
  if (target.kind === "meld" && !targetMeld) return null;

  const moving = new Set(ids);
  let melds = layout.melds.map((m) => m.filter((t) => !moving.has(t)));
  let pos = [...layout.pos];
  const hand = layout.hand.filter((t) => !moving.has(t));

  if (target.kind === "hand") {
    hand.push(...ids);
    return prune({ melds, pos, hand });
  }
  if (target.kind === "new") {
    melds.push([...ids]);
    pos.push(target.at);
    const l = prune({ melds, pos, hand });
    return settle(l, l.melds.length - 1, target.at);
  }

  const index = layout.melds.indexOf(targetMeld!);
  const meld = melds[index];
  const at = target.before === null ? meld.length : meld.indexOf(target.before);
  meld.splice(at < 0 ? meld.length : at, 0, ...ids);
  // A longer meld may now touch its neighbour: nudge it to a free spot.
  ({ melds, pos } = prune({ melds, pos, hand }));
  const newIndex = melds.indexOf(meld);
  return settle({ melds, pos, hand }, newIndex, pos[newIndex]);
}

/** Moves a whole meld to another cell of the table (RN32). */
export function moveMeld(layout: Layout, index: number, at: MeldPos): Layout | null {
  if (!layout.melds[index]) return null;
  const next = settle(layout, index, at);
  const [x, y] = next.pos[index];
  const [ox, oy] = layout.pos[index];
  return x === ox && y === oy ? null : next;
}

/** Joins melds[from] to the start or the end of melds[to] (RN34). */
export function joinMelds(
  layout: Layout,
  from: number,
  to: number,
  side: "start" | "end",
): Layout | null {
  if (from === to || !layout.melds[from] || !layout.melds[to]) return null;
  const joined =
    side === "end"
      ? [...layout.melds[to], ...layout.melds[from]]
      : [...layout.melds[from], ...layout.melds[to]];
  const melds = layout.melds.map((m, i) => (i === to ? joined : i === from ? [] : m));
  // Joining at the start keeps the right end where it was.
  const [tx, ty] = layout.pos[to];
  const pos = [...layout.pos];
  if (side === "start") pos[to] = [tx - TILE_COLS * layout.melds[from].length, ty];
  const l = prune({ melds, pos, hand: layout.hand });
  const index = l.melds.indexOf(joined);
  return settle(l, index, l.pos[index]);
}

/** Cuts melds[index] in two before its tile number `at` (RN33). */
export function cutMeld(layout: Layout, index: number, at: number): Layout | null {
  const meld = layout.melds[index];
  if (!meld || at <= 0 || at >= meld.length) return null;
  const left = meld.slice(0, at);
  const right = meld.slice(at);
  const [x, y] = layout.pos[index];
  const melds = [...layout.melds];
  melds[index] = left;
  melds.push(right);
  const pos: MeldPos[] = [...layout.pos, [x + meldCols(left.length) + 1, y]];
  return settle({ melds, pos, hand: layout.hand }, melds.length - 1, pos[pos.length - 1]);
}

/** Positions for melds that came without them, in rows of three (as the server's AutoPos). */
export function autoPos(n: number): MeldPos[] {
  return Array.from({ length: n }, (_, i) => [
    (i % 3) * (TABLE_COLS / 3),
    (Math.floor(i / 3) * 3) % TABLE_ROWS,
  ]);
}
