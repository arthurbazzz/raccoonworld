import type { Graphics } from "pixi.js";
import { describe, expect, it } from "vitest";

import { DEFAULT_PALETTE, outlineOf, shade, type RaccoonPalette } from "./palette";
import { buildFront } from "./rig";

function fills(g: Graphics): number[] {
  return g.context.instructions
    .filter((i) => i.action === "fill")
    .map((i) => (i.data as { style: { color: number } }).style.color);
}

describe("front pose (RW-003/T5)", () => {
  const parts = () => buildFront(DEFAULT_PALETTE);

  it("stacks the layers back to front: tail, legs, body, arms, head (RA2)", () => {
    const p = parts();
    expect(p.body.children).toEqual([p.tail, p.legLayer, p.torso, p.armLayer, p.head]);
  });

  it("draws the face as the top layer of the head, so it moves with it", () => {
    const p = parts();
    expect(p.face.parent).toBe(p.head);
    expect(p.head.children.at(-1)).toBe(p.face);
  });

  it("has two legs and two arms", () => {
    const p = parts();
    expect(p.legs).toHaveLength(2);
    expect(p.arms).toHaveLength(2);
    expect(p.legLayer.children).toEqual(p.legs);
    expect(p.armLayer.children).toEqual(p.arms);
  });

  it("has a big head, about half of the height (reference proportions)", () => {
    const p = parts();
    const all = p.body.getLocalBounds();
    const head = p.head.getBounds(); // includes the ears
    const height = all.maxY - all.minY;
    expect(height).toBeGreaterThanOrEqual(40);
    expect(height).toBeLessThanOrEqual(48);
    expect((head.maxY - head.minY) / height).toBeGreaterThan(0.5);
  });

  it("keeps the feet on the ground line and centered (RA1)", () => {
    const b = parts().body.getLocalBounds();
    expect(b.maxY).toBeLessThanOrEqual(2);
    expect(b.maxY).toBeGreaterThan(-1);
    expect(Math.abs(b.minX + b.maxX) / 2).toBeLessThan(6);
  });

  it("shows the mask, the cream band, eyes with shine and a nose on the face", () => {
    const f = fills(parts().face);
    expect(f).toContain(DEFAULT_PALETTE.mask);
    expect(f).toContain(DEFAULT_PALETTE.furLight);
    expect(f).toContain(DEFAULT_PALETTE.eye);
    expect(f).toContain(DEFAULT_PALETTE.nose);
  });

  it("paints with the palette it is given, shadows included (RA5, RA7)", () => {
    const brown: RaccoonPalette = { ...DEFAULT_PALETTE, fur: 0x8b5a2b };
    const torso = fills(buildFront(brown).torso);
    expect(torso[0]).toBe(shade(brown.fur));
    expect(torso).toContain(brown.fur);
    const strokes = buildFront(brown).torso.context.instructions.filter(
      (i) => i.action === "stroke",
    );
    expect((strokes[0].data as { style: { color: number } }).style.color).toBe(outlineOf(brown));
  });
});
