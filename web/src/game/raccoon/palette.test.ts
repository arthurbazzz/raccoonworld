import { describe, expect, it } from "vitest";

import {
  DEFAULT_PALETTE,
  OUTLINE_FACTOR,
  outlineOf,
  shade,
  SHADE_FACTOR,
  shadeOf,
  type RaccoonPalette,
} from "./palette";

const rgb = (c: number) => [(c >> 16) & 0xff, (c >> 8) & 0xff, c & 0xff];

describe("shade", () => {
  it("darkens each channel by the factor, rounding", () => {
    expect(shade(0xffffff, 0.5)).toBe(0x808080);
    expect(rgb(shade(0x8090a0, 0.75))).toEqual([96, 108, 120]);
  });

  it("keeps black black and never goes out of range", () => {
    expect(shade(0x000000, 0.75)).toBe(0x000000);
    expect(shade(0xffffff, 1)).toBe(0xffffff);
    expect(shade(0xffffff, 0)).toBe(0x000000);
  });

  it("darkens by 25% by default (RA7)", () => {
    expect(SHADE_FACTOR).toBe(0.75);
    expect(shade(0xffffff)).toBe(shade(0xffffff, SHADE_FACTOR));
  });
});

describe("derived tones (RA5, RA7)", () => {
  it("derives every shadow from its base color", () => {
    const s = shadeOf(DEFAULT_PALETTE);
    for (const key of Object.keys(DEFAULT_PALETTE) as (keyof RaccoonPalette)[]) {
      expect(s[key]).toBe(shade(DEFAULT_PALETTE[key]));
    }
  });

  it("derives the outline from the fur, darker than its shadow", () => {
    expect(OUTLINE_FACTOR).toBeLessThan(SHADE_FACTOR);
    expect(outlineOf(DEFAULT_PALETTE)).toBe(shade(DEFAULT_PALETTE.fur, OUTLINE_FACTOR));
  });

  it("changing only the fur changes its shadow and the outline too", () => {
    const brown: RaccoonPalette = { ...DEFAULT_PALETTE, fur: 0x8b5a2b };
    expect(shadeOf(brown).fur).not.toBe(shadeOf(DEFAULT_PALETTE).fur);
    expect(outlineOf(brown)).not.toBe(outlineOf(DEFAULT_PALETTE));
    expect(shadeOf(brown).mask).toBe(shadeOf(DEFAULT_PALETTE).mask);
  });
});

describe("DEFAULT_PALETTE", () => {
  it("has the colors of the reference art", () => {
    expect(Object.keys(DEFAULT_PALETTE).sort()).toEqual(
      ["earInner", "eye", "fur", "furLight", "mask", "nose", "tailRing"].sort(),
    );
  });

  it("keeps the mask darker than the fur and the belly lighter (readable at 1x, RA6)", () => {
    const lum = (c: number) => rgb(c).reduce((a, b) => a + b, 0);
    expect(lum(DEFAULT_PALETTE.mask)).toBeLessThan(lum(DEFAULT_PALETTE.fur));
    expect(lum(DEFAULT_PALETTE.furLight)).toBeGreaterThan(lum(DEFAULT_PALETTE.fur));
  });
});
