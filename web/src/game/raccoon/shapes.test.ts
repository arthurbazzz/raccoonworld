import { Graphics } from "pixi.js";
import { describe, expect, it } from "vitest";

import { DEFAULT_PALETTE, outlineOf, shade } from "./palette";
import {
  highlightOf,
  OUTLINE_WIDTH,
  pointedEar,
  ringedTail,
  shadedCircle,
  shadedEllipse,
} from "./shapes";

const P = DEFAULT_PALETTE;
const OUTLINE = outlineOf(P);

/** The fills and strokes of a Graphics, in drawing order. */
function ops(g: Graphics): { action: string; color: number | undefined }[] {
  return g.context.instructions.map((i) => ({
    action: i.action,
    color: (i.data as { style?: { color?: number } }).style?.color,
  }));
}

describe("shadedEllipse (RA7: two-shape shading)", () => {
  it("draws the shadow tone with the outline, then the base color on top", () => {
    const g = shadedEllipse(new Graphics(), 0, 0, 10, 8, P.fur, OUTLINE);
    expect(ops(g)).toEqual([
      { action: "fill", color: shade(P.fur) },
      { action: "stroke", color: OUTLINE },
      { action: "fill", color: P.fur },
    ]);
  });

  it("puts the base shape inside, shifted up and left, so the shadow shows bottom-right", () => {
    const h = highlightOf(0, 0, 10, 8);
    expect(h.x).toBeLessThan(0);
    expect(h.y).toBeLessThan(0);
    expect(h.rx).toBeLessThan(10);
    expect(h.ry).toBeLessThan(8);
    // Still inside the shadow shape on every side.
    expect(h.x - h.rx).toBeGreaterThanOrEqual(-10);
    expect(h.y - h.ry).toBeGreaterThanOrEqual(-8);
  });

  it("can skip the outline for inner details", () => {
    const g = shadedEllipse(new Graphics(), 0, 0, 4, 3, P.furLight, null);
    expect(ops(g).map((o) => o.action)).toEqual(["fill", "fill"]);
  });

  it("uses a thin outline that stays visible at game size", () => {
    expect(OUTLINE_WIDTH).toBeGreaterThanOrEqual(1);
    expect(OUTLINE_WIDTH).toBeLessThanOrEqual(1.5);
  });
});

describe("shadedCircle", () => {
  it("is a shaded ellipse with equal radii", () => {
    const a = ops(shadedCircle(new Graphics(), 0, 0, 6, P.fur, OUTLINE));
    const b = ops(shadedEllipse(new Graphics(), 0, 0, 6, 6, P.fur, OUTLINE));
    expect(a).toEqual(b);
  });
});

describe("pointedEar", () => {
  it("is a shaded, outlined point with a pink inner ear", () => {
    const g = pointedEar(new Graphics(), 0, 0, 8, 10, P, OUTLINE);
    const colors = ops(g);
    expect(colors).toContainEqual({ action: "stroke", color: OUTLINE });
    expect(colors).toContainEqual({ action: "fill", color: P.earInner });
    // The inner ear is drawn last, on top.
    expect(colors.at(-1)).toEqual({ action: "fill", color: P.earInner });
  });

  it("points up from its base at (x, y)", () => {
    const b = pointedEar(new Graphics(), 0, 0, 8, 10, P, OUTLINE).getLocalBounds();
    expect(b.maxY).toBeLessThanOrEqual(OUTLINE_WIDTH);
    expect(b.minY).toBeLessThan(-9);
  });
});

describe("ringedTail", () => {
  it("grows up from its base at the origin, so rotating it sways from the root", () => {
    const b = ringedTail(new Graphics(), 5, 9, P, OUTLINE).getLocalBounds();
    expect(b.maxY).toBeLessThanOrEqual(OUTLINE_WIDTH);
    expect(b.minY).toBeLessThan(-17);
    expect((b.minX + b.maxX) / 2).toBeCloseTo(0, 0);
  });

  it("is shaded fur with dark rings", () => {
    const colors = ops(ringedTail(new Graphics(), 5, 9, P, OUTLINE));
    expect(colors.slice(0, 3)).toEqual([
      { action: "fill", color: shade(P.fur) },
      { action: "stroke", color: OUTLINE },
      { action: "fill", color: P.fur },
    ]);
    expect(colors.filter((o) => o.color === P.tailRing).length).toBeGreaterThanOrEqual(3);
  });
});
