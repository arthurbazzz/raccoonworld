// Moving one tile in our turn's arrangement (RN11): pure, so the drag and
// drop in the panel only has to say where the tile was dropped.

export interface Layout {
  melds: number[][];
  hand: number[];
}

/** Where a tile is dropped: into a meld (before a tile, or at its end), a new meld, or the hand. */
export type DropTarget =
  { kind: "meld"; index: number; before: number | null } | { kind: "new" } | { kind: "hand" };

/**
 * Moves `id` to `target`. Melds left empty disappear. A tile that was on the
 * table at the start of the turn cannot go back to the hand: `startHand`
 * says which tiles may (the server would refuse it anyway, RN11).
 * Returns null when the move changes nothing or is not allowed.
 */
export function moveTile(
  layout: Layout,
  id: number,
  target: DropTarget,
  startHand: number[],
): Layout | null {
  if (target.kind === "hand" && !startHand.includes(id)) return null;
  if (target.kind === "meld" && target.before === id) return null;

  const fromHand = layout.hand.includes(id);
  const fromMeld = layout.melds.findIndex((m) => m.includes(id));
  if (!fromHand && fromMeld < 0) return null;
  if (target.kind === "hand" && fromHand) return null;

  // The target meld, before removing anything, so its index stays right.
  const targetMeld = target.kind === "meld" ? layout.melds[target.index] : null;
  if (target.kind === "meld" && !targetMeld) return null;

  let melds = layout.melds.map((m) => m.filter((t) => t !== id));
  const hand = layout.hand.filter((t) => t !== id);

  if (target.kind === "hand") {
    hand.push(id);
  } else if (target.kind === "new") {
    melds.push([id]);
  } else {
    const meld = melds[target.index];
    const at = target.before === null ? meld.length : meld.indexOf(target.before);
    meld.splice(at < 0 ? meld.length : at, 0, id);
  }
  melds = melds.filter((m) => m.length > 0);
  return { melds, hand };
}
