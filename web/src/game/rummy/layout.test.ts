import { describe, expect, it } from "vitest";

import {
  autoPos,
  cutMeld,
  freeSpot,
  joinMelds,
  meldCols,
  moveMeld,
  moveTiles,
  type Layout,
} from "./layout";

const layout: Layout = {
  melds: [
    [1, 2, 3],
    [20, 21, 22, 23],
  ],
  pos: [
    [0, 0],
    [20, 0],
  ],
  hand: [7, 8, 9],
};
const start = [7, 8, 9];

describe("moveTiles (RN11, RN33, RN35)", () => {
  it("lays hand tiles on a new meld where they are dropped", () => {
    const next = moveTiles(layout, [7, 8], { kind: "new", at: [10, 10] }, start)!;
    expect(next.melds[2]).toEqual([7, 8]);
    expect(next.pos[2]).toEqual([10, 10]);
    expect(next.hand).toEqual([9]);
  });

  it("puts tiles before another one, or at the end of a meld", () => {
    expect(moveTiles(layout, [7], { kind: "meld", index: 0, before: 2 }, start)?.melds[0]).toEqual([
      1, 7, 2, 3,
    ]);
    expect(
      moveTiles(layout, [8, 9], { kind: "meld", index: 1, before: null }, start)?.melds[1],
    ).toEqual([20, 21, 22, 23, 8, 9]);
  });

  it("takes tiles out of a meld into a new one", () => {
    const next = moveTiles(layout, [22, 23], { kind: "new", at: [0, 10] }, start)!;
    expect(next.melds).toEqual([
      [1, 2, 3],
      [20, 21],
      [22, 23],
    ]);
  });

  it("drops a meld left empty, with its position", () => {
    const next = moveTiles(layout, [1, 2, 3], { kind: "meld", index: 1, before: null }, start)!;
    expect(next.melds).toEqual([[20, 21, 22, 23, 1, 2, 3]]);
    expect(next.pos).toHaveLength(1);
  });

  it("only lets hand tiles go back to the hand", () => {
    const laid = moveTiles(layout, [7], { kind: "new", at: [0, 10] }, start)!;
    expect(moveTiles(laid, [7], { kind: "hand" }, start)?.hand).toEqual([8, 9, 7]);
    expect(moveTiles(layout, [1], { kind: "hand" }, start)).toBeNull();
  });

  it("refuses moves that change nothing or name unknown tiles", () => {
    expect(moveTiles(layout, [7], { kind: "hand" }, start)).toBeNull();
    expect(moveTiles(layout, [2], { kind: "meld", index: 0, before: 2 }, start)).toBeNull();
    expect(moveTiles(layout, [99], { kind: "new", at: [0, 0] }, start)).toBeNull();
    expect(moveTiles(layout, [], { kind: "new", at: [0, 0] }, start)).toBeNull();
  });

  it("never drops a new meld on top of another", () => {
    const next = moveTiles(layout, [7], { kind: "new", at: [1, 0] }, start)!;
    const [x, y] = next.pos[2];
    const overlaps = y < 3 && x < meldCols(3) + 1;
    expect(overlaps).toBe(false);
  });
});

describe("moveMeld (RN32)", () => {
  it("moves a meld to a free cell", () => {
    expect(moveMeld(layout, 0, [5, 12])?.pos[0]).toEqual([5, 12]);
  });

  it("keeps the meld inside the table", () => {
    expect(moveMeld(layout, 0, [100, 100])?.pos[0]).toEqual([48 - meldCols(3), 25]);
  });

  it("returns null when the meld stays where it was", () => {
    expect(moveMeld(layout, 0, [0, 0])).toBeNull();
  });
});

describe("cutMeld and joinMelds (RN33, RN34)", () => {
  it("cuts a meld in two, the right part beside the left one", () => {
    const next = cutMeld(layout, 1, 2)!;
    expect(next.melds).toEqual([
      [1, 2, 3],
      [20, 21],
      [22, 23],
    ]);
    expect(next.pos[2][0]).toBeGreaterThan(next.pos[1][0]);
  });

  it("does not cut at either end", () => {
    expect(cutMeld(layout, 1, 0)).toBeNull();
    expect(cutMeld(layout, 1, 4)).toBeNull();
  });

  it("joins a meld at the end or the start of another", () => {
    expect(joinMelds(layout, 1, 0, "end")?.melds).toEqual([[1, 2, 3, 20, 21, 22, 23]]);
    expect(joinMelds(layout, 0, 1, "start")?.melds).toEqual([[1, 2, 3, 20, 21, 22, 23]]);
    expect(joinMelds(layout, 0, 0, "end")).toBeNull();
  });
});

describe("freeSpot and autoPos", () => {
  it("finds the nearest free cell", () => {
    const melds = [[1, 2, 3], [4]];
    const spot = freeSpot(
      melds,
      [
        [0, 0],
        [0, 0],
      ],
      1,
      [0, 0],
    );
    expect(spot).not.toEqual([0, 0]);
  });

  it("lays melds in rows of three, like the server", () => {
    expect(autoPos(4)).toEqual([
      [0, 0],
      [16, 0],
      [32, 0],
      [0, 3],
    ]);
  });
});
