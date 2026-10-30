// The tile rack (RN30): two rows of slots the player arranges at will. It is
// local only; the server never sees where a tile sits on the rack.

import { byColor, byNumber } from "./tiles";

export const RACK_ROWS = 2;
export const MIN_SLOTS_PER_ROW = 14;

/** Slots, row by row; null is an empty slot. */
export type Rack = (number | null)[];

/** Slots per row for a hand: at least 14, and room for every tile plus a few gaps. */
export const slotsPerRow = (handSize: number) =>
  Math.max(MIN_SLOTS_PER_ROW, Math.ceil((handSize + 4) / RACK_ROWS));

/** Fits the rack to `hand`: tiles that left free their slot, new tiles take the first free one. */
export function syncRack(rack: Rack, hand: number[]): Rack {
  const want = new Set(hand);
  const per = slotsPerRow(hand.length);
  const size = per * RACK_ROWS;
  let next: Rack = rack.map((t) => (t !== null && want.has(t) ? t : null));
  // Keep each row's slots in place when the rack grows: widen row by row.
  const oldPer = next.length / RACK_ROWS;
  if (next.length !== size && Number.isInteger(oldPer) && oldPer <= per) {
    const rows: Rack = [];
    for (let r = 0; r < RACK_ROWS; r++) {
      const row = next.slice(r * oldPer, (r + 1) * oldPer);
      rows.push(...row, ...Array<null>(per - oldPer).fill(null));
    }
    next = rows;
  } else if (next.length !== size) {
    next = next.filter((t) => t !== null);
  }
  while (next.length < size) next.push(null);
  const placed = new Set(next.filter((t): t is number => t !== null));
  for (const id of hand) {
    if (placed.has(id)) continue;
    const free = next.indexOf(null);
    if (free < 0) next.push(id);
    else next[free] = id;
  }
  return next;
}

/** Lays the hand out sorted, left to right, row by row (Por cor / Por número). */
export function sortRack(hand: number[], by: "color" | "number"): Rack {
  const sorted = [...hand].sort(by === "color" ? byColor : byNumber);
  const size = slotsPerRow(hand.length) * RACK_ROWS;
  return [...sorted, ...Array<null>(size - sorted.length).fill(null)];
}

/**
 * Puts tiles `ids` into the rack starting at `slot`, the first one there and
 * the others in the next free slots after it. Tiles already on the rack
 * leave their old slots first.
 */
export function placeOnRack(rack: Rack, ids: number[], slot: number): Rack {
  const next = rack.map((t) => (t !== null && ids.includes(t) ? null : t));
  let at = slot;
  for (const id of ids) {
    while (at < next.length && next[at] !== null) at++;
    if (at >= next.length) {
      const free = next.indexOf(null);
      if (free < 0) next.push(id);
      else next[free] = id;
    } else {
      next[at] = id;
    }
  }
  return next;
}
