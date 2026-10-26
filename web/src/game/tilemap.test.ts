import { describe, expect, it } from "vitest";

import { fitScale, mapUrl, parseTileMap } from "./tilemap";

const json = {
  name: "test",
  tileSize: 32,
  spawn: { x: 48, y: 48 },
  rows: ["####", "#..#", "####"],
};

describe("parseTileMap", () => {
  it("reads size and solid tiles", () => {
    const m = parseTileMap(json);
    expect(m).toMatchObject({
      name: "test",
      cols: 4,
      rows: 3,
      tileSize: 32,
      width: 128,
      height: 96,
    });
    expect(m.isSolid(0, 0)).toBe(true);
    expect(m.isSolid(1, 1)).toBe(false);
    expect(m.isSolid(-1, 1)).toBe(true); // outside is solid, like the server
  });

  it("lists solid tiles with their pixel rect", () => {
    const m = parseTileMap({ ...json, rows: ["#.", ".#"] });
    expect(m.solidTiles()).toEqual([
      { col: 0, row: 0, x: 0, y: 0, size: 32 },
      { col: 1, row: 1, x: 32, y: 32, size: 32 },
    ]);
  });

  it.each([
    ["no rows", { ...json, rows: [] }],
    ["ragged rows", { ...json, rows: ["###", "##"] }],
    ["unknown tile", { ...json, rows: ["#?#"] }],
    ["bad tile size", { ...json, tileSize: 0 }],
  ])("rejects %s", (_name, bad) => {
    expect(() => parseTileMap(bad)).toThrow();
  });
});

describe("fitScale", () => {
  it("fits the map inside the view keeping proportions", () => {
    expect(fitScale(960, 640, 960, 640)).toBe(1);
    expect(fitScale(480, 640, 960, 640)).toBe(0.5); // width-bound
    expect(fitScale(1920, 640, 960, 640)).toBe(1); // height-bound
  });

  it("never scales above 2x", () => {
    expect(fitScale(4000, 4000, 960, 640)).toBe(2);
  });
});

describe("mapUrl", () => {
  it("derives the HTTP map URL from the WebSocket URL", () => {
    expect(mapUrl("ws://localhost:8080/ws", "praca")).toBe("http://localhost:8080/maps/praca");
    expect(mapUrl("wss://raccoons.world/ws", "praca")).toBe("https://raccoons.world/maps/praca");
  });
});

describe("booth tiles (RW-005)", () => {
  const m = parseTileMap({ ...json, rows: ["######", "#.BB.#", "######"] });

  it("are solid, like on the server", () => {
    expect(m.isSolid(2, 1)).toBe(true);
    expect(m.isSolid(3, 1)).toBe(true);
  });

  it("form the booth box and are not drawn as wall blocks", () => {
    expect(m.booth).toEqual({ x: 64, y: 32, w: 64, h: 32 });
    expect(m.solidTiles().some((t) => t.row === 1 && (t.col === 2 || t.col === 3))).toBe(false);
  });

  it("are absent from maps without 'B'", () => {
    expect(parseTileMap(json).booth).toBeNull();
  });
});
