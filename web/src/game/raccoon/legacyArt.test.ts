import { describe, expect, it } from "vitest";

import { RACCOON_COLORS, RaccoonView } from "./RaccoonView";

// Details of the RW-002 drawing. They are replaced by RW-003 tests as the
// new poses (T5-T7) and palette (T3) land; behavior tests live in
// RaccoonView.test.ts and must keep passing.
describe("RW-002 drawing details (replaced in RW-003)", () => {
  it("shows two paws from the front and back and four from the side", () => {
    const v = new RaccoonView();
    expect(v.parts.legs).toHaveLength(2);
    v.setFacing("up");
    expect(v.parts.legs).toHaveLength(2);
    v.setFacing("right");
    expect(v.parts.legs).toHaveLength(4);
  });

  it("uses a single palette", () => {
    expect(Object.keys(RACCOON_COLORS)).toEqual(
      expect.arrayContaining(["fur", "furLight", "mask", "tailRing", "paw"]),
    );
  });
});
