import { describe, expect, it } from "vitest";

import { moveTile } from "./layout";

const layout = {
  melds: [
    [1, 2, 3],
    [20, 21, 22, 23],
  ],
  hand: [7, 8],
};
const start = [7, 8];

describe("moveTile (RN11)", () => {
  it("lays a hand tile on a new meld", () => {
    expect(moveTile(layout, 7, { kind: "new" }, start)).toEqual({
      melds: [[1, 2, 3], [20, 21, 22, 23], [7]],
      hand: [8],
    });
  });

  it("puts a tile before another one, or at the end of a meld", () => {
    expect(moveTile(layout, 7, { kind: "meld", index: 0, before: 2 }, start)?.melds[0]).toEqual([
      1, 7, 2, 3,
    ]);
    expect(moveTile(layout, 8, { kind: "meld", index: 1, before: null }, start)?.melds[1]).toEqual([
      20, 21, 22, 23, 8,
    ]);
  });

  it("moves a table tile between melds and drops the empty ones", () => {
    const one = { melds: [[5], [1, 2, 3]], hand: [] };
    expect(moveTile(one, 5, { kind: "meld", index: 1, before: null }, [])).toEqual({
      melds: [[1, 2, 3, 5]],
      hand: [],
    });
  });

  it("takes a tile from a meld to a new one", () => {
    expect(moveTile(layout, 23, { kind: "new" }, start)?.melds).toEqual([
      [1, 2, 3],
      [20, 21, 22],
      [23],
    ]);
  });

  it("brings our own tile back to the hand, but never a table tile", () => {
    const laid = moveTile(layout, 7, { kind: "new" }, start)!;
    expect(moveTile(laid, 7, { kind: "hand" }, start)).toEqual({
      melds: layout.melds,
      hand: [8, 7],
    });
    expect(moveTile(layout, 1, { kind: "hand" }, start)).toBeNull();
  });

  it("ignores moves that change nothing", () => {
    expect(moveTile(layout, 7, { kind: "hand" }, start)).toBeNull();
    expect(moveTile(layout, 2, { kind: "meld", index: 0, before: 2 }, start)).toBeNull();
    expect(moveTile(layout, 99, { kind: "new" }, start)).toBeNull();
  });
});
