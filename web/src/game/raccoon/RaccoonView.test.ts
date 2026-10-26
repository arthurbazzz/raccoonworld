import { describe, expect, it } from "vitest";

import {
  animationFrame,
  BREATH_PERIOD_MS,
  RaccoonView,
  STEPS_PER_SECOND,
} from "./RaccoonView";

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
});

describe("animationFrame (RV2)", () => {
  it("stands still and breathes when idle", () => {
    expect(animationFrame(false, 0)).toEqual({ bob: 0, legLift: [0, 0], tailSway: 0, breath: 1 });
    const peak = animationFrame(false, BREATH_PERIOD_MS / 2);
    expect(peak.breath).toBeCloseTo(1.03);
    expect(peak.bob).toBe(0);
    expect(peak.legLift).toEqual([0, 0]);
  });

  it("breathes between 1.00 and 1.03 over a ~2s cycle", () => {
    expect(BREATH_PERIOD_MS).toBe(2000);
    for (let t = 0; t < 4000; t += 37) {
      const b = animationFrame(false, t).breath;
      expect(b).toBeGreaterThanOrEqual(1);
      expect(b).toBeLessThanOrEqual(1.03 + 1e-9);
    }
  });

  it("bobs up to 2px and alternates paws while walking", () => {
    let sawA = false;
    let sawB = false;
    for (let t = 0; t < 1000; t += 7) {
      const f = animationFrame(true, t);
      expect(f.bob).toBeLessThanOrEqual(0);
      expect(f.bob).toBeGreaterThanOrEqual(-2);
      expect(f.breath).toBe(1);
      // Never both pairs in the air.
      expect(Math.min(f.legLift[0], f.legLift[1])).toBe(0);
      if (f.legLift[0] > 0) sawA = true;
      if (f.legLift[1] > 0) sawB = true;
    }
    expect(sawA && sawB).toBe(true);
  });

  it("takes about 6 steps per second", () => {
    expect(STEPS_PER_SECOND).toBe(6);
    // One step later the other pair of paws is up.
    const stepMs = 1000 / STEPS_PER_SECOND;
    const a = animationFrame(true, stepMs / 2);
    const b = animationFrame(true, stepMs * 1.5);
    expect(a.legLift[0]).toBeGreaterThan(0);
    expect(b.legLift[1]).toBeGreaterThan(0);
  });

  it("sways the tail while walking", () => {
    const sways = [0, 40, 80, 120].map((t) => animationFrame(true, t).tailSway);
    expect(new Set(sways.map((s) => s.toFixed(3))).size).toBeGreaterThan(1);
  });
});

describe("RaccoonView.animate", () => {
  it("applies the walking frame to the parts and resets when idle", () => {
    const v = new RaccoonView();
    const legY = v.parts.legs.map((l) => l.y);
    const tailRotation = v.parts.tail.rotation;

    const stepMs = 1000 / STEPS_PER_SECOND;
    v.animate(true, stepMs / 2);
    expect(v.parts.body.y).toBeLessThan(0);
    expect(v.parts.legs[0].y).toBeLessThan(legY[0]);
    expect(v.parts.tail.rotation).not.toBe(tailRotation);

    v.animate(false, 0);
    expect(v.parts.body.y).toBe(0);
    expect(v.parts.legs.map((l) => l.y)).toEqual(legY);
    expect(v.parts.tail.rotation).toBe(tailRotation);
  });

  it("breathes by stretching the torso upward and lifting the head with it", () => {
    const v = new RaccoonView();
    const headY = v.parts.head.y;
    v.animate(false, BREATH_PERIOD_MS / 2);
    expect(v.parts.torso.scale.y).toBeCloseTo(1.03);
    expect(v.parts.head.y).toBeLessThan(headY);
  });

  it("keeps animating after the pose changes", () => {
    const v = new RaccoonView();
    v.setFacing("right");
    v.animate(true, 1000 / STEPS_PER_SECOND / 2);
    expect(v.parts.body.y).toBeLessThan(0);
  });
});

describe("ring under your own raccoon (RV3)", () => {
  it("has no ring by default", () => {
    expect(new RaccoonView().ring).toBeNull();
  });

  it("draws a ring at the feet, below the body and the shadow", () => {
    const v = new RaccoonView({ isYou: true });
    expect(v.ring).not.toBeNull();
    expect(v.getChildIndex(v.ring!)).toBe(0);
    const b = v.ring!.getLocalBounds();
    expect((b.minX + b.maxX) / 2).toBeCloseTo(0);
    expect((b.minY + b.maxY) / 2).toBeCloseTo(0);
    expect(b.maxX - b.minX).toBeGreaterThan(20); // wider than the 20px hitbox
  });

  it("stays put while the raccoon turns and walks (RV4)", () => {
    const v = new RaccoonView({ isYou: true });
    const ring = v.ring!;
    v.setFacing("left");
    v.animate(true, 1000 / STEPS_PER_SECOND / 2);
    expect(v.ring).toBe(ring);
    expect(ring.scale.x).toBe(1);
    expect(ring.position.x).toBe(0);
    expect(ring.position.y).toBe(0);
    expect(v.getChildIndex(ring)).toBe(0);
  });
});
