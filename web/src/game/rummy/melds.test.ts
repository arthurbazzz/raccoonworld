import { readFileSync } from "node:fs";
import path from "node:path";

import { describe, expect, it } from "vitest";

import { classify, points } from "./melds";
import { byColor, byNumber, DECK_SIZE, faceOf, tileName, tileOf } from "./tiles";

// The same cases as the Go domain's TestClassify (plan, section 6).
interface MeldCase {
  name: string;
  tiles: string[];
  kind?: string;
  points?: number;
  error?: string;
}

const cases = JSON.parse(
  readFileSync(
    path.resolve(__dirname, "../../../../server/internal/domain/rummy/testdata/melds.json"),
    "utf8",
  ),
) as MeldCase[];

/** IDs for faces, taking a fresh copy for each repeat. */
function ids(faces: string[]): number[] {
  const used = new Set<number>();
  return faces.map((f) => {
    for (let id = 0; id < DECK_SIZE; id++) {
      if (!used.has(id) && faceOf(id) === f) {
        used.add(id);
        return id;
      }
    }
    throw new Error(`no tile ${f}`);
  });
}

describe("classify (shared cases with the server)", () => {
  it.each(cases.map((c) => [c.name, c] as const))("%s", (_, c) => {
    const tiles = ids(c.tiles);
    if (c.error) {
      expect(classify(tiles)).toEqual({ error: c.error });
      expect(points(tiles)).toBe(0);
    } else {
      expect(classify(tiles)).toEqual({ kind: c.kind });
      expect(points(tiles)).toBe(c.points);
    }
  });
});

describe("tiles", () => {
  it("follow the server layout", () => {
    expect(faceOf(0)).toBe("brown-1");
    expect(faceOf(51)).toBe("blue-13");
    expect(faceOf(52)).toBe("brown-1");
    expect(faceOf(104)).toBe("joker");
    expect(tileOf(105).joker).toBe(true);
    expect(tileName(13 + 6)).toBe("7 vermelho");
    expect(tileName(105)).toBe("coringa");
  });

  it("sort by color or by number, jokers last", () => {
    const hand = ids(["blue-1", "joker", "brown-9", "brown-2", "red-2"]);
    expect([...hand].sort(byColor).map(faceOf)).toEqual([
      "brown-2",
      "brown-9",
      "red-2",
      "blue-1",
      "joker",
    ]);
    expect([...hand].sort(byNumber).map(faceOf)).toEqual([
      "blue-1",
      "brown-2",
      "red-2",
      "brown-9",
      "joker",
    ]);
  });
});
