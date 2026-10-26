import { describe, expect, it } from "vitest";

import { findFrames, SpriteSheetError, type RGBAImage } from "./slice.ts";

/** A transparent image. */
function blank(width: number, height: number): RGBAImage {
  return { width, height, data: new Uint8Array(width * height * 4) };
}

/** Paints an opaque (or given alpha) rectangle. */
function rect(img: RGBAImage, x: number, y: number, w: number, h: number, alpha = 255): RGBAImage {
  for (let yy = y; yy < y + h; yy++) {
    for (let xx = x; xx < x + w; xx++) {
      const i = (yy * img.width + xx) * 4;
      img.data[i] = 120;
      img.data[i + 1] = 120;
      img.data[i + 2] = 130;
      img.data[i + 3] = alpha;
    }
  }
  return img;
}

/** 2 rows x 3 columns of 20x30 "raccoons", 40px apart. */
function grid(): RGBAImage {
  const img = blank(130, 100);
  for (let r = 0; r < 2; r++) for (let c = 0; c < 3; c++) rect(img, 5 + c * 42, 5 + r * 48, 20, 30);
  return img;
}

describe("findFrames (RS2)", () => {
  it("finds rows and columns from the transparent gaps", () => {
    const frames = findFrames(grid(), { rows: 2, cols: 3 });
    expect(frames).toHaveLength(2);
    expect(frames[0]).toHaveLength(3);
    expect(frames[1][2]).toEqual({ x0: 89, x1: 108, top: 53, bottom: 82 });
  });

  it("does not need a regular grid: rows can be anywhere", () => {
    const img = blank(130, 120);
    for (let c = 0; c < 3; c++) {
      rect(img, 5 + c * 42, 2, 20, 40); // row 1 nearly touches the top
      rect(img, 5 + c * 42, 61, 20, 50); // row 2 starts mid-image
    }
    const frames = findFrames(img, { rows: 2, cols: 3 });
    expect(frames[0][0]).toMatchObject({ top: 2, bottom: 41 });
    expect(frames[1][0]).toMatchObject({ top: 61, bottom: 110 });
  });

  it("measures each frame tightly, so feet at different heights are kept", () => {
    const img = blank(130, 60);
    rect(img, 5, 5, 20, 30); // feet at y=34
    rect(img, 47, 8, 20, 30); // feet at y=37
    rect(img, 89, 5, 20, 31); // feet at y=35
    const [row] = findFrames(img, { rows: 1, cols: 3 });
    expect(row.map((f) => f.bottom)).toEqual([34, 37, 35]);
  });

  it("ignores specks smaller than minSize (AI noise)", () => {
    const img = grid();
    rect(img, 60, 42, 2, 2); // between the rows
    rect(img, 30, 15, 2, 2); // between two columns
    expect(() => findFrames(img, { rows: 2, cols: 3 })).not.toThrow();
  });

  it("ignores nearly transparent pixels (alpha below the threshold)", () => {
    const img = grid();
    rect(img, 0, 40, 130, 4, 60); // a faint haze across the gap
    expect(findFrames(img, { rows: 2, cols: 3 })).toHaveLength(2);
  });

  it("fails with the counts it found when a frame is missing", () => {
    const img = blank(130, 100);
    for (let c = 0; c < 3; c++) rect(img, 5 + c * 42, 5, 20, 30);
    for (let c = 0; c < 2; c++) rect(img, 5 + c * 42, 53, 20, 30);
    expect(() => findFrames(img, { rows: 2, cols: 3 })).toThrow(SpriteSheetError);
    expect(() => findFrames(img, { rows: 2, cols: 3 })).toThrow(
      "esperado 2×3 quadros, achei 2 fileiras com 3, 2 colunas",
    );
  });

  it("fails when the number of rows is wrong", () => {
    expect(() => findFrames(grid(), { rows: 3, cols: 3 })).toThrow(
      "esperado 3×3 quadros, achei 2 fileiras com 3, 3 colunas",
    );
  });

  it("fails on an empty image", () => {
    expect(() => findFrames(blank(10, 10), { rows: 1, cols: 1 })).toThrow("achei 0 fileiras");
  });
});
