// Client view of a room tile map. The JSON comes from the server
// (GET /maps/:name), the same file it uses for collision.

export interface TileMapJSON {
  name: string;
  tileSize: number;
  spawn: { x: number; y: number };
  rows: string[];
}

export interface SolidTile {
  col: number;
  row: number;
  /** Top-left corner in room pixels. */
  x: number;
  y: number;
  size: number;
}

export interface TileMap {
  name: string;
  cols: number;
  rows: number;
  tileSize: number;
  width: number;
  height: number;
  isSolid(col: number, row: number): boolean;
  solidTiles(): SolidTile[];
}

export function parseTileMap(raw: TileMapJSON): TileMap {
  const { name, tileSize, rows } = raw;
  if (!(tileSize > 0)) throw new Error("tileSize must be positive");
  if (rows.length === 0 || rows[0].length === 0) throw new Error("map has no tiles");
  const cols = rows[0].length;

  const solid: boolean[] = [];
  rows.forEach((row, r) => {
    if (row.length !== cols) throw new Error(`row ${r} has ${row.length} tiles, want ${cols}`);
    for (const ch of row) {
      if (ch !== "#" && ch !== ".")
        throw new Error(`unknown tile ${JSON.stringify(ch)} at row ${r}`);
      solid.push(ch === "#");
    }
  });

  const isSolid = (col: number, row: number) =>
    col < 0 || row < 0 || col >= cols || row >= rows.length || solid[row * cols + col];

  return {
    name,
    cols,
    rows: rows.length,
    tileSize,
    width: cols * tileSize,
    height: rows.length * tileSize,
    isSolid,
    solidTiles: () => {
      const out: SolidTile[] = [];
      for (let row = 0; row < rows.length; row++) {
        for (let col = 0; col < cols; col++) {
          if (solid[row * cols + col]) {
            out.push({ col, row, x: col * tileSize, y: row * tileSize, size: tileSize });
          }
        }
      }
      return out;
    },
  };
}

const MAX_SCALE = 2;

/** Largest uniform scale that fits the map in the view, capped at 2x. */
export function fitScale(viewW: number, viewH: number, mapW: number, mapH: number): number {
  return Math.min(viewW / mapW, viewH / mapH, MAX_SCALE);
}

/** HTTP URL of a map on the same server as the WebSocket endpoint. */
export function mapUrl(wsUrl: string, name: string): string {
  const u = new URL(wsUrl);
  u.protocol = u.protocol === "wss:" ? "https:" : "http:";
  u.pathname = `/maps/${encodeURIComponent(name)}`;
  return u.toString();
}

export async function loadTileMap(wsUrl: string, name: string): Promise<TileMap> {
  const res = await fetch(mapUrl(wsUrl, name));
  if (!res.ok) throw new Error(`map ${name}: HTTP ${res.status}`);
  return parseTileMap((await res.json()) as TileMapJSON);
}
