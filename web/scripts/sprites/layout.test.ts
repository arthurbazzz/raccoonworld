import { describe, expect, it } from "vitest";

import { layoutSheet, manifestFor, PAD } from "./layout.ts";
import type { FrameBox } from "./slice.ts";

// Two rows of two frames with different sizes and feet heights, like the AI art.
const frames: FrameBox[][] = [
  [
    { x0: 10, x1: 49, top: 5, bottom: 54 }, // 40x50
    { x0: 70, x1: 103, top: 8, bottom: 59 }, // 34x52
  ],
  [
    { x0: 12, x1: 47, top: 80, bottom: 127 }, // 36x48
    { x0: 71, x1: 106, top: 79, bottom: 130 }, // 36x52
  ],
];

describe("layoutSheet (RS3)", () => {
  const layout = () => layoutSheet(frames);

  it("makes every cell as big as the biggest frame plus padding", () => {
    expect(layout().cell).toEqual({ w: 40 + 2 * PAD, h: 52 + 2 * PAD });
  });

  it("puts the anchor at the bottom center of the cell, under the feet", () => {
    const { cell, anchor } = layout();
    expect(anchor).toEqual({ x: Math.floor(cell.w / 2), y: cell.h - PAD });
  });

  it("places every frame with its feet on the anchor of its cell", () => {
    const { cell, anchor, placements } = layout();
    for (const p of placements) {
      const w = p.src.x1 - p.src.x0 + 1;
      const h = p.src.bottom - p.src.top + 1;
      // Feet line: the row right under the lowest pixel.
      expect(p.y + h).toBe(p.row * cell.h + anchor.y);
      // Feet center within half a pixel of the anchor.
      expect(Math.abs(p.x + w / 2 - (p.col * cell.w + anchor.x))).toBeLessThanOrEqual(0.5);
    }
  });

  it("keeps every frame inside its own cell", () => {
    const { cell, placements } = layout();
    for (const p of placements) {
      const w = p.src.x1 - p.src.x0 + 1;
      const h = p.src.bottom - p.src.top + 1;
      expect(p.x).toBeGreaterThanOrEqual(p.col * cell.w);
      expect(p.x + w).toBeLessThanOrEqual((p.col + 1) * cell.w);
      expect(p.y).toBeGreaterThanOrEqual(p.row * cell.h);
      expect(p.y + h).toBeLessThanOrEqual((p.row + 1) * cell.h);
    }
  });

  it("uses the median frame height as the art height (for scaling)", () => {
    expect(layout().artHeight).toBe(51); // heights 50, 52, 48, 52 -> median 51
  });

  it("sizes the output sheet to the grid of cells", () => {
    const { cell, size } = layout();
    expect(size).toEqual({ w: 2 * cell.w, h: 2 * cell.h });
  });
});

describe("manifestFor", () => {
  it("describes the agreed layout: rows front/back/side, 2 idle and 4 walk columns", () => {
    const layout = layoutSheet(frames);
    expect(manifestFor(layout, "/sprites/raccoon.png")).toEqual({
      image: "/sprites/raccoon.png",
      cell: layout.cell,
      anchor: layout.anchor,
      rows: { front: 0, back: 1, side: 2 },
      idle: [0, 1],
      walk: [2, 3, 4, 5],
      artHeight: 51,
    });
  });
});
