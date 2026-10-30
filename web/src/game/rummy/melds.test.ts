import { readFileSync } from "node:fs";
import path from "node:path";

import { describe, expect, it } from "vitest";

import { checkDraft, classify, describeIssue, diagnose, points } from "./melds";
import { byColor, byNumber, DECK_SIZE, faceOf, tileName, tileOf } from "./tiles";

// The same cases as the Go domain's TestClassify (plan, section 6).
interface MeldCase {
  name: string;
  tiles: string[];
  kind?: string;
  points?: number;
  error?: string;
  culprits?: number[];
  params?: Record<string, number>;
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
      expect(diagnose(tiles)).toEqual({
        issue: { code: c.error, tiles: (c.culprits ?? []).map((i) => tiles[i]), params: c.params },
      });
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

describe("describeIssue (RN12a)", () => {
  const say = (faces: string[]) => {
    const d = diagnose(ids(faces));
    if (!("issue" in d)) throw new Error("valid meld");
    return describeIssue(d.issue);
  };

  it("names the repeated color", () => {
    expect(say(["green-7", "green-7", "blue-7", "red-7"])).toBe(
      "O grupo de 7 tem dois 7 verdes. Tire um deles.",
    );
  });

  it("says how many tiles a short meld has", () => {
    expect(say(["blue-7", "green-7"])).toBe(
      "Esta combinação tem só 2 peças; precisa de pelo menos 3.",
    );
  });

  it("points at the tile of the wrong color", () => {
    expect(say(["blue-3", "red-4", "blue-5"])).toBe(
      "Sequência com cores diferentes: o 4 vermelho não combina com as azuis.",
    );
  });

  it("says which number is missing", () => {
    expect(say(["blue-3", "blue-4", "blue-6"])).toBe(
      "Depois do 4 azul vem o 5, não o 6. Falta o 5 azul.",
    );
  });

  it("explains a run past 13", () => {
    expect(say(["red-12", "red-13", "joker"])).toBe(
      "A sequência passou do 13. Tire o coringa do fim.",
    );
  });

  it("explains the opening", () => {
    expect(
      describeIssue({ code: "opening_too_low", tiles: [], params: { points: 24, need: 6 } }),
    ).toBe("Sua abertura soma 24 pontos; faltam 6 para 30.");
  });
});

describe("checkDraft", () => {
  it("lists every invalid meld with its index", () => {
    const issues = checkDraft(
      [ids(["blue-3", "blue-4", "blue-5"]), ids(["red-7", "blue-7"])],
      [],
      [],
      false,
    );
    expect(issues.map((i) => [i.meld, i.code])).toEqual([[1, "too_short"]]);
  });

  it("checks the opening points", () => {
    const run = ids(["blue-3", "blue-4", "blue-5"]);
    const issues = checkDraft([run], [], run, true);
    expect(issues).toEqual([
      { meld: -1, code: "opening_too_low", tiles: [], params: { points: 12, need: 18 } },
    ]);
  });

  it("blames table tiles used in the opening", () => {
    const table = ids(["red-10", "blue-10", "green-10"]);
    const ten = ids(["brown-10"]);
    const issues = checkDraft([[...table, ...ten]], [table], ten, true);
    expect(issues[0]).toMatchObject({ meld: 0, code: "opening_used_table" });
  });
});
