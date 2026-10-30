// Marking several tiles to move them together (RN35).

/** Marks or unmarks a tile (Shift+click). */
export function toggle(selected: number[], id: number): number[] {
  return selected.includes(id) ? selected.filter((t) => t !== id) : [...selected, id];
}

export interface Box {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

/** A box from two corners, in any order. */
export function boxFrom(ax: number, ay: number, bx: number, by: number): Box {
  return {
    left: Math.min(ax, bx),
    top: Math.min(ay, by),
    right: Math.max(ax, bx),
    bottom: Math.max(ay, by),
  };
}

/** Tiles whose box overlaps the selection box (the rectangle on the table). */
export function tilesInBox(box: Box, tiles: { id: number; rect: Box }[]): number[] {
  return tiles
    .filter(
      ({ rect: r }) =>
        r.left < box.right && box.left < r.right && r.top < box.bottom && box.top < r.bottom,
    )
    .map((t) => t.id);
}

/**
 * The tiles a drag of `id` carries, in the order they lie in `order` (the
 * table melds, then the rack): all the marked ones if `id` is marked, or
 * just `id`.
 */
export function dragged(selected: number[], id: number, order: number[]): number[] {
  if (!selected.includes(id)) return [id];
  return order.filter((t) => selected.includes(t));
}
