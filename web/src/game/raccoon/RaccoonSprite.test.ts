import { TextureSource } from "pixi.js";
import { describe, expect, it } from "vitest";

import { IDLE_FRAME_MS, WALK_FPS } from "./frames";
import type { SheetManifest } from "./manifest";
import { RaccoonView, TARGET_HEIGHT } from "./RaccoonSprite";
import { sheetFrom, type RaccoonSheet } from "./sheet";

const manifest: SheetManifest = {
  image: "/sprites/raccoon.png",
  cell: { w: 109, h: 120 },
  anchor: { x: 54, y: 116 },
  rows: { front: 0, back: 1, side: 2 },
  idle: [0, 1],
  walk: [2, 3, 4, 5],
  artHeight: 108,
};

function fakeSheet(): RaccoonSheet {
  const source = new TextureSource({ width: 6 * 109, height: 3 * 120 });
  return sheetFrom(manifest, source);
}

describe("sheetFrom", () => {
  it("cuts one texture per cell of the aligned sheet", () => {
    const s = fakeSheet();
    expect(s.frames).toHaveLength(3);
    expect(s.frames[0]).toHaveLength(6);
    const f = s.frames[2][4].frame;
    expect([f.x, f.y, f.width, f.height]).toEqual([4 * 109, 2 * 120, 109, 120]);
  });
});

describe("RaccoonView in sprites (RW-004)", () => {
  it("stands on its feet: the sprite anchor is the feet point of the cell (RS3)", () => {
    const v = new RaccoonView(fakeSheet());
    expect(v.sprite.anchor.x).toBeCloseTo(54 / 109);
    expect(v.sprite.anchor.y).toBeCloseTo(116 / 120);
  });

  it("is about 46px tall in the room (RS4)", () => {
    expect(TARGET_HEIGHT).toBe(46);
    const v = new RaccoonView(fakeSheet());
    expect(v.sprite.scale.y * manifest.artHeight).toBeCloseTo(46);
  });

  it("shows the frame chosen by facing, state and time", () => {
    const s = fakeSheet();
    const v = new RaccoonView(s);
    v.setFacing("up");
    v.animate(false, 0);
    expect(v.sprite.texture).toBe(s.frames[1][0]);
    v.animate(true, (1000 / WALK_FPS) * 2 + 1);
    expect(v.sprite.texture).toBe(s.frames[1][4]);
    v.setFacing("down");
    v.animate(false, IDLE_FRAME_MS + 1);
    expect(v.sprite.texture).toBe(s.frames[0][1]);
  });

  it("mirrors only the sprite when facing left (RW-002)", () => {
    const s = fakeSheet();
    const v = new RaccoonView(s, { isYou: true });
    v.setFacing("left");
    v.animate(false, 0);
    expect(v.sprite.texture).toBe(s.frames[2][0]);
    expect(v.sprite.scale.x).toBeLessThan(0);
    expect(v.scale.x).toBe(1);
    expect(v.ring!.scale.x).toBe(1);
    v.setFacing("right");
    v.animate(false, 0);
    expect(v.sprite.scale.x).toBeGreaterThan(0);
  });

  it("turns right away, without waiting for the next frame", () => {
    const s = fakeSheet();
    const v = new RaccoonView(s);
    v.animate(true, 0);
    v.setFacing("up");
    expect(v.sprite.texture).toBe(s.frames[1][2]);
  });

  it("has the ring only for your own raccoon, under the sprite (RW-002 RV3)", () => {
    expect(new RaccoonView(fakeSheet()).ring).toBeNull();
    const you = new RaccoonView(fakeSheet(), { isYou: true });
    expect(you.ring).not.toBeNull();
    expect(you.getChildIndex(you.ring!)).toBeLessThan(you.getChildIndex(you.sprite));
  });

  it("starts facing down, idle", () => {
    const s = fakeSheet();
    const v = new RaccoonView(s);
    expect(v.facing).toBe("down");
    expect(v.sprite.texture).toBe(s.frames[0][0]);
  });
});
