import { describe, expect, it } from "vitest";

import type { CupsRoundData } from "@/net/protocol";

import { cupsAt, finalNutSlot, shuffleEndMs } from "./timeline";

const round: CupsRoundData = {
  round: 1,
  total: 5,
  nut: 0,
  showMs: 1000,
  swaps: [
    [0, 1],
    [1, 2],
  ],
  swapMs: 500,
  pickMs: 10_000,
  elapsedMs: 0,
};

const slots = (t: number) => cupsAt(round, t).cups.map((c) => c.slot);

describe("cupsAt", () => {
  it("shows the nut: only its cup goes up, then down again", () => {
    expect(cupsAt(round, 0).stage).toBe("show");
    expect(cupsAt(round, 500).cups.map((c) => c.lift)).toEqual([1, 0, 0]);
    expect(cupsAt(round, 0).cups[0].lift).toBe(0);
    expect(cupsAt(round, 999).cups[0].lift).toBeLessThan(0.01);
  });

  it("moves the two swapped cups between their slots, one in front, one behind", () => {
    const f = cupsAt(round, 1250); // halfway through the first swap
    expect(f.stage).toBe("shuffle");
    expect(f.cups[0].slot).toBeCloseTo(0.5);
    expect(f.cups[1].slot).toBeCloseTo(0.5);
    expect(f.cups[0].arc).toBeCloseTo(1);
    expect(f.cups[1].arc).toBeCloseTo(-1);
    expect(f.cups[2]).toEqual({ slot: 2, arc: 0, lift: 0 });
  });

  it("chains the swaps: the second one moves whoever is in those slots now", () => {
    expect(slots(1500)).toEqual([1, 0, 2]); // first swap done
    expect(slots(2000)).toEqual([2, 0, 1]); // cup 0 was in slot 1, now in 2
  });

  it("lets players pick once the cups stop", () => {
    expect(shuffleEndMs(round)).toBe(2000);
    expect(cupsAt(round, 1999).stage).toBe("shuffle");
    expect(cupsAt(round, 2000).stage).toBe("choose");
    expect(cupsAt(round, 60_000).stage).toBe("choose");
  });

  it("depends only on the time, so a tab coming back jumps to the right place", () => {
    expect(cupsAt(round, 1750)).toEqual(cupsAt(round, 1750));
    expect(slots(1500)).toEqual([1, 0, 2]);
  });
});

describe("finalNutSlot (RC17)", () => {
  it("follows the nut through every swap", () => {
    expect(finalNutSlot(round)).toBe(2);
    expect(finalNutSlot({ ...round, nut: 2 })).toBe(1);
    expect(finalNutSlot({ ...round, swaps: [] })).toBe(0);
  });

  it("matches where the nut's cup is drawn at the end", () => {
    const f = cupsAt(round, shuffleEndMs(round));
    expect(f.cups[f.nutCup].slot).toBe(finalNutSlot(round));
  });
});
