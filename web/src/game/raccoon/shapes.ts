import type { Graphics } from "pixi.js";

import { shade, type RaccoonPalette } from "./palette";

// Shapes with outline and simple shading (RW-003, RA7). Every part is drawn
// twice: the whole shape in its shadow tone with the outline, then the base
// color slightly smaller and shifted up-left. The uncovered sliver at the
// bottom-right is the shadow, as if lit from the top-left.

/** Outline thickness in room pixels: about 1px at game size. */
export const OUTLINE_WIDTH = 1.2;

const HIGHLIGHT_SCALE = 0.86;
const HIGHLIGHT_SHIFT_X = 0.1;
const HIGHLIGHT_SHIFT_Y = 0.12;

/** Where the base-color shape goes inside a shadow ellipse. */
export function highlightOf(x: number, y: number, rx: number, ry: number) {
  return {
    x: x - rx * HIGHLIGHT_SHIFT_X,
    y: y - ry * HIGHLIGHT_SHIFT_Y,
    rx: rx * HIGHLIGHT_SCALE,
    ry: ry * HIGHLIGHT_SCALE,
  };
}

function strokeStyle(color: number) {
  return { color, width: OUTLINE_WIDTH, join: "round" as const, cap: "round" as const };
}

/**
 * A shaded ellipse. Pass `outline = null` for inner details (belly, muzzle)
 * that sit inside an already outlined part.
 */
export function shadedEllipse(
  g: Graphics,
  x: number,
  y: number,
  rx: number,
  ry: number,
  base: number,
  outline: number | null,
): Graphics {
  g.ellipse(x, y, rx, ry).fill(shade(base));
  if (outline !== null) g.stroke(strokeStyle(outline));
  const h = highlightOf(x, y, rx, ry);
  return g.ellipse(h.x, h.y, h.rx, h.ry).fill(base);
}

export function shadedCircle(
  g: Graphics,
  x: number,
  y: number,
  r: number,
  base: number,
  outline: number | null,
): Graphics {
  return shadedEllipse(g, x, y, r, r, base, outline);
}

/**
 * A pointed ear growing up from its base at (x, y). `tilt` moves the tip
 * sideways, so the left and right ears can lean outward.
 */
export function pointedEar(
  g: Graphics,
  x: number,
  y: number,
  w: number,
  h: number,
  palette: RaccoonPalette,
  outline: number,
  tilt = 0,
): Graphics {
  // Puffy sides that meet at a soft point, like the reference (not a flat triangle).
  const leaf = (halfW: number, top: number, bottom: number, tipX: number) =>
    g
      .moveTo(x - halfW, bottom)
      .quadraticCurveTo(x - halfW * 1.1 + tilt * 0.3, bottom - (bottom - top) * 0.65, tipX, top)
      .quadraticCurveTo(
        x + halfW * 1.1 + tilt * 0.3,
        bottom - (bottom - top) * 0.65,
        x + halfW,
        bottom,
      )
      .closePath();

  leaf(w / 2, y - h, y, x + tilt)
    .fill(shade(palette.fur))
    .stroke(strokeStyle(outline));
  leaf(w * 0.4, y - h * 0.9, y, x + tilt * 0.9 - w * 0.06).fill(palette.fur);
  return leaf(w * 0.22, y - h * 0.66, y - h * 0.1, x + tilt * 0.7).fill(palette.earInner);
}

/**
 * A bushy tail with three dark rings and a dark tip, growing up from its
 * base at the origin. Position and rotate the Graphics to place it, so it
 * sways from the root.
 */
export function ringedTail(
  g: Graphics,
  rx: number,
  ry: number,
  palette: RaccoonPalette,
  outline: number,
): Graphics {
  shadedEllipse(g, 0, -ry, rx, ry, palette.fur, outline);
  for (const k of [-0.45, 0, 0.45]) {
    g.ellipse(0, -ry + k * ry, rx * 0.92, ry * 0.11).fill(palette.tailRing);
  }
  return g.ellipse(0, -ry * 1.8, rx * 0.55, ry * 0.2).fill(palette.tailRing);
}
