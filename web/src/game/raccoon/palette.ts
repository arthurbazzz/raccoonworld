// The raccoon's colors (RW-003, RA5 and RA7). Only base colors are stored;
// shadows and the outline are derived, so recoloring the fur recolors its
// shading and outline too. A future customization spec swaps the palette.

export interface RaccoonPalette {
  fur: number;
  furLight: number; // belly, muzzle and the band above the mask
  mask: number;
  earInner: number;
  tailRing: number;
  nose: number;
  eye: number; // the eye shine
}

/** Colors picked from the reference sprite sheet. */
export const DEFAULT_PALETTE: RaccoonPalette = {
  fur: 0x8a8790,
  furLight: 0xe9e2d6,
  mask: 0x26242b,
  earInner: 0xd99a96,
  tailRing: 0x3d3a42,
  nose: 0x1d1b20,
  eye: 0xffffff,
};

/** How much darker the shaded side of a part is (25%). */
export const SHADE_FACTOR = 0.75;
/** The outline is the fur at about a third of its brightness. */
export const OUTLINE_FACTOR = 0.34;

/** Multiplies each RGB channel by `factor` (0..1), rounding. */
export function shade(color: number, factor: number = SHADE_FACTOR): number {
  const k = Math.min(Math.max(factor, 0), 1);
  const r = Math.round(((color >> 16) & 0xff) * k);
  const g = Math.round(((color >> 8) & 0xff) * k);
  const b = Math.round((color & 0xff) * k);
  return (r << 16) | (g << 8) | b;
}

/** The shadow tone of every palette color. */
export function shadeOf(palette: RaccoonPalette): RaccoonPalette {
  const out = { ...palette };
  for (const key of Object.keys(out) as (keyof RaccoonPalette)[]) out[key] = shade(palette[key]);
  return out;
}

/** The outline color drawn around every part. */
export function outlineOf(palette: RaccoonPalette): number {
  return shade(palette.fur, OUTLINE_FACTOR);
}
