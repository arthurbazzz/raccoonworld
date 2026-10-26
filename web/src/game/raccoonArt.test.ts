import { describe, expect, it } from "vitest";

import { RACCOON_COLORS, RaccoonView } from "./raccoonArt";

describe("RaccoonView poses (RV1)", () => {
  it("starts facing down with the front pose", () => {
    const v = new RaccoonView();
    expect(v.facing).toBe("down");
    expect(v.pose).toBe("front");
  });

  it.each([
    ["down", "front", 1],
    ["up", "back", 1],
    ["right", "side", 1],
    ["left", "side", -1],
  ] as const)("facing %s uses the %s pose (scale.x %d)", (facing, pose, scaleX) => {
    const v = new RaccoonView();
    v.setFacing(facing);
    expect(v.pose).toBe(pose);
    expect(v.parts.body.scale.x).toBe(scaleX);
  });

  it("mirrors only the body: things added next to it are not flipped", () => {
    const v = new RaccoonView();
    const label = v.addChild(new RaccoonView()); // stand-in for the nickname
    v.setFacing("left");
    expect(v.parts.body.scale.x).toBe(-1);
    expect(v.scale.x).toBe(1);
    expect(label.scale.x).toBe(1);
  });

  it("turning between left and right only flips, without rebuilding the pose", () => {
    const v = new RaccoonView();
    v.setFacing("right");
    const body = v.parts.body;
    v.setFacing("left");
    expect(v.parts.body).toBe(body);
  });

  it("rebuilds the parts when the pose changes", () => {
    const v = new RaccoonView();
    const front = v.parts.body;
    v.setFacing("up");
    expect(v.parts.body).not.toBe(front);
    expect(front.destroyed).toBe(true);
  });

  it("draws the tail behind the body from the front and in front of it from the back", () => {
    const v = new RaccoonView();
    const order = () => v.parts.body.children;
    expect(order().indexOf(v.parts.tail)).toBeLessThan(order().indexOf(v.parts.torso));
    v.setFacing("up");
    expect(order().indexOf(v.parts.tail)).toBeGreaterThan(order().indexOf(v.parts.torso));
  });

  it("shows two paws from the front and back and four from the side", () => {
    const v = new RaccoonView();
    expect(v.parts.legs).toHaveLength(2);
    v.setFacing("up");
    expect(v.parts.legs).toHaveLength(2);
    v.setFacing("right");
    expect(v.parts.legs).toHaveLength(4);
  });

  it("keeps the feet at the origin and the art above them (RV4)", () => {
    for (const facing of ["down", "up", "right", "left"] as const) {
      const v = new RaccoonView();
      v.setFacing(facing);
      const b = v.parts.body.getLocalBounds();
      expect(b.maxY).toBeLessThanOrEqual(2); // paws may touch the ground line
      expect(b.minY).toBeGreaterThan(-45); // fits under the nickname at y=-38..-50
      expect(Math.abs(b.minX + b.maxX) / 2).toBeLessThan(8); // roughly centered on the feet
    }
  });

  it("uses a single palette", () => {
    expect(Object.keys(RACCOON_COLORS)).toEqual(
      expect.arrayContaining(["fur", "furLight", "mask", "tailRing", "paw"]),
    );
  });
});
