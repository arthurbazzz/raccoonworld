import { describe, expect, it } from "vitest";

import { rummyManifest, splitStrip } from "./rummy.ts";
import { SpriteSheetError, type RGBAImage } from "./slice.ts";

function blank(width: number, height: number): RGBAImage {
  return { width, height, data: new Uint8Array(width * height * 4) };
}

function rect(img: RGBAImage, x: number, y: number, w: number, h: number): void {
  for (let yy = y; yy < y + h; yy++) {
    for (let xx = x; xx < x + w; xx++) img.data[(yy * img.width + xx) * 4 + 3] = 255;
  }
}

/** Touching tiles of width w, with a dip of `dip` transparent pixels at each seam. */
function strip(img: RGBAImage, x: number, y: number, widths: number[], h: number, dip = 4): void {
  let left = x;
  widths.forEach((w, i) => {
    rect(img, left, y, w, h);
    if (i > 0) {
      // Rounded corners where two tiles meet: the seam column is shorter.
      for (let d = 0; d < dip; d++) {
        img.data[((y + d) * img.width + left) * 4 + 3] = 0;
        img.data[((y + h - 1 - d) * img.width + left) * 4 + 3] = 0;
      }
    }
    left += w;
  });
}

describe("splitStrip (RW-006)", () => {
  it("cuts touching tiles at their seams, the last one wider", () => {
    const img = blank(400, 40);
    strip(img, 10, 5, [20, 20, 20, 40], 30);
    const pieces = splitStrip(img, { x0: 10, x1: 109, top: 5, bottom: 34 }, 4);
    expect(pieces.map((p) => [p.x0, p.x1])).toEqual([
      [10, 29],
      [30, 49],
      [50, 69],
      [70, 109],
    ]);
  });

  it("fails loudly when the count does not match", () => {
    const img = blank(400, 40);
    strip(img, 10, 5, [20, 20, 20], 30);
    expect(() => splitStrip(img, { x0: 10, x1: 69, top: 5, bottom: 34 }, 4)).toThrow(
      SpriteSheetError,
    );
  });
});

describe("rummyManifest", () => {
  it("finds 13 tiles per color, the joker, the brown back, the stack and the table", () => {
    const img = blank(500, 260);
    for (let c = 0; c < 4; c++) {
      const y = 5 + c * 40;
      rect(img, 5, y, 20, 30); // back
      rect(img, 35, y, 25, 30); // stack
      strip(img, 70, y, [...Array(13).fill(20), 40], 30);
    }
    rect(img, 5, 175, 100, 70); // table
    rect(img, 120, 175, 20, 30);
    rect(img, 150, 175, 25, 30);

    const m = rummyManifest(img, "/sprites/rummy.png");
    expect(m.size).toEqual({ w: 500, h: 260 });
    expect(m.tiles).toHaveLength(4);
    expect(m.tiles.every((row) => row.length === 13)).toBe(true);
    expect(m.tiles[2][0]).toEqual({ x: 70, y: 85, w: 20, h: 30 });
    expect(m.joker).toEqual({ x: 330, y: 5, w: 40, h: 30 });
    expect(m.back).toEqual({ x: 5, y: 5, w: 20, h: 30 });
    expect(m.stack).toEqual({ x: 35, y: 5, w: 25, h: 30 });
    expect(m.table).toEqual({ x: 5, y: 175, w: 100, h: 70 });
  });
});
