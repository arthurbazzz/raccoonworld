import { describe, expect, it } from "vitest";

import { placeOnRack, slotsPerRow, sortRack, syncRack } from "./rack";
import { boxFrom, dragged, tilesInBox, toggle } from "./selection";

describe("rack (RN30)", () => {
  it("has two rows of at least 14 slots", () => {
    expect(slotsPerRow(14)).toBe(14);
    expect(slotsPerRow(30)).toBe(17);
    expect(syncRack([], [1, 2, 3])).toHaveLength(28);
  });

  it("keeps tiles where the player put them, and frees slots of tiles laid down", () => {
    let rack = placeOnRack(syncRack([], [1, 2, 3]), [1], 20);
    expect(rack[20]).toBe(1);
    rack = syncRack(rack, [1, 3]);
    expect(rack[20]).toBe(1);
    expect(rack).not.toContain(2);
  });

  it("puts a drawn tile in the first free slot", () => {
    const rack = placeOnRack(syncRack([], [1, 2]), [1], 5);
    const next = syncRack(rack, [1, 2, 9]);
    expect(next.indexOf(9)).toBe(0); // the slot 1 left
  });

  it("widens each row in place when the hand grows", () => {
    let rack = syncRack([], [1]);
    rack = placeOnRack(rack, [1], 14); // first slot of the bottom row
    const hand = Array.from({ length: 26 }, (_, i) => i + 1);
    rack = syncRack(rack, hand);
    expect(rack.length).toBe(2 * slotsPerRow(26));
    expect(rack[slotsPerRow(26)]).toBe(1);
  });

  it("sorts by color or by number", () => {
    // 0 = brown 1, 13 = red 1, 1 = brown 2
    expect(sortRack([13, 1, 0], "color").slice(0, 3)).toEqual([0, 1, 13]);
    expect(sortRack([13, 1, 0], "number").slice(0, 3)).toEqual([0, 13, 1]);
  });

  it("moves several tiles together into the next free slots", () => {
    const rack = syncRack([], [1, 2, 3]);
    const next = placeOnRack(rack, [1, 2], 10);
    expect(next.slice(10, 12)).toEqual([1, 2]);
  });
});

describe("selection (RN35)", () => {
  it("marks and unmarks tiles", () => {
    expect(toggle(toggle([], 5), 7)).toEqual([5, 7]);
    expect(toggle([5, 7], 5)).toEqual([7]);
  });

  it("finds the tiles inside a rectangle", () => {
    const box = boxFrom(100, 100, 0, 0);
    const tiles = [
      { id: 1, rect: { left: 10, top: 10, right: 30, bottom: 40 } },
      { id: 2, rect: { left: 200, top: 10, right: 230, bottom: 40 } },
    ];
    expect(tilesInBox(box, tiles)).toEqual([1]);
  });

  it("drags all marked tiles in table order, or just the one grabbed", () => {
    expect(dragged([9, 3], 3, [1, 3, 5, 9])).toEqual([3, 9]);
    expect(dragged([9, 3], 5, [1, 3, 5, 9])).toEqual([5]);
  });
});
