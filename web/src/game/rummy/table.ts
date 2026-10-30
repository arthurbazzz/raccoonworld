import { BOOTH_REACH, nearBooth } from "../cups/booth";
import type { Rect } from "../tilemap";

/** How close the feet must be to the pieces table to play (RN4): same as the booth. */
export const TABLE_REACH = BOOTH_REACH;

/** Whether feet at (x, y) are within TABLE_REACH of the table (Go `TileMap.NearTable`). */
export function nearTable(table: Rect | null, x: number, y: number): boolean {
  return nearBooth(table, x, y);
}
