import { Container, Graphics } from "pixi.js";

import type { Facing } from "./motion";

// The raccoon drawn in code (RW-002, Q1). Everything is built upward from
// the feet at (0, 0), so depth sorting and the hitbox are unchanged (RV4).

/** The single raccoon palette; a future customization spec swaps it. */
export const RACCOON_COLORS = {
  fur: 0x8d8d94,
  furDark: 0x6c6c74,
  furLight: 0xe6e1d8,
  mask: 0x2d2d33,
  nose: 0x1c1c21,
  eye: 0xffffff,
  pupil: 0x141418,
  earInner: 0xc7a39b,
  tailRing: 0x3b3b42,
  paw: 0x4a4a52,
  shadow: 0x000000,
};

export type Pose = "front" | "back" | "side";

/** The animatable parts of a pose. */
export interface RaccoonParts {
  /** Everything that flips when facing left. */
  body: Container;
  torso: Graphics;
  head: Container;
  tail: Graphics;
  legs: Graphics[];
}

const C = RACCOON_COLORS;

export function poseFor(facing: Facing): Pose {
  if (facing === "down") return "front";
  if (facing === "up") return "back";
  return "side";
}

/**
 * A raccoon in one of three poses: front (facing down), back (facing up)
 * and side (facing right; left is the same pose mirrored). Only `body`
 * flips, so a nickname or ring added to the view stays readable.
 */
export class RaccoonView extends Container {
  facing: Facing = "down";
  pose: Pose = "front";
  parts: RaccoonParts;

  constructor() {
    super();
    this.addChild(drawShadow());
    this.parts = buildPose(this.pose);
    this.addChild(this.parts.body);
  }

  setFacing(facing: Facing): void {
    this.facing = facing;
    const pose = poseFor(facing);
    if (pose !== this.pose) {
      const index = this.getChildIndex(this.parts.body);
      this.parts.body.destroy({ children: true });
      this.parts = buildPose(pose);
      this.addChildAt(this.parts.body, index);
      this.pose = pose;
    }
    this.parts.body.scale.x = facing === "left" ? -1 : 1;
  }
}

function drawShadow(): Graphics {
  return new Graphics().ellipse(0, 0, 11, 4).fill({ color: C.shadow, alpha: 0.22 });
}

function buildPose(pose: Pose): RaccoonParts {
  switch (pose) {
    case "front":
      return buildFront();
    case "back":
      return buildBack();
    case "side":
      return buildSide();
  }
}

function paw(x: number, color: number = C.paw): Graphics {
  return new Graphics().ellipse(x, -1.5, 3.2, 2.2).fill(color);
}

/**
 * A bushy tail with dark rings, growing upward from its base at (baseX,
 * baseY). The base is the Graphics origin, so rotating it sways the tail
 * from the root.
 */
function stripedTail(baseX: number, baseY: number, rx: number, ry: number): Graphics {
  const g = new Graphics().ellipse(0, -ry, rx, ry).fill(C.fur);
  for (const k of [-0.45, 0, 0.45]) {
    g.ellipse(0, -ry + k * ry, rx * 0.95, ry * 0.12).fill(C.tailRing);
  }
  g.ellipse(0, -ry * 1.85, rx * 0.6, ry * 0.18).fill(C.tailRing); // dark tip
  g.position.set(baseX, baseY);
  return g;
}

function ears(g: Graphics, dx: number, y: number): Graphics {
  for (const x of [-dx, dx]) {
    g.circle(x, y, 4).fill(C.fur);
    g.circle(x, y + 0.5, 2).fill(C.earInner);
  }
  return g;
}

function buildFront(): RaccoonParts {
  const body = new Container();
  const tail = stripedTail(7, -5, 4.5, 8);
  tail.rotation = 0.35;
  const legs = [paw(-5), paw(5)];
  const torso = new Graphics()
    .ellipse(0, -11, 10, 10)
    .fill(C.fur)
    .ellipse(0, -9, 6, 7)
    .fill(C.furLight);

  const head = new Container();
  head.position.set(0, -25);
  const face = ears(new Graphics(), 7, -8)
    .circle(0, 0, 10)
    .fill(C.fur)
    .ellipse(0, 4, 7, 5)
    .fill(C.furLight)
    .roundRect(-9.5, -4, 19, 6, 3)
    .fill(C.mask);
  for (const x of [-4, 4]) face.circle(x, -1, 2).fill(C.eye).circle(x, -0.8, 1).fill(C.pupil);
  face.ellipse(0, 3.5, 1.8, 1.3).fill(C.nose);
  head.addChild(face);

  body.addChild(tail, ...legs, torso, head);
  return { body, torso, head, tail, legs };
}

function buildBack(): RaccoonParts {
  const body = new Container();
  const legs = [paw(-5, C.furDark), paw(5, C.furDark)];
  const torso = new Graphics().ellipse(0, -11, 10, 10).fill(C.fur);

  const head = new Container();
  head.position.set(0, -25);
  head.addChild(
    ears(new Graphics(), 7, -8).circle(0, 0, 10).fill(C.fur).ellipse(0, -2, 6, 4).fill(C.furDark),
  );

  // From behind, the ringed tail is the star of the pose, but leaning to the
  // side so the back and head still read as a raccoon.
  const tail = stripedTail(3, -3, 4.5, 8);
  tail.rotation = 0.45;
  body.addChild(...legs, torso, head, tail);
  return { body, torso, head, tail, legs };
}

function buildSide(): RaccoonParts {
  const body = new Container();
  const tail = stripedTail(-8, -7, 4.5, 9);
  tail.rotation = -0.9;
  const farLegs = [paw(-4, C.furDark), paw(6, C.furDark)];
  const nearLegs = [paw(-6), paw(4)];
  const torso = new Graphics()
    .ellipse(0, -11, 11, 9)
    .fill(C.fur)
    .ellipse(2, -8, 6, 5)
    .fill(C.furLight);

  const head = new Container();
  head.position.set(5, -24);
  const face = new Graphics()
    .circle(-2, -8, 4)
    .fill(C.fur)
    .circle(-2, -7.5, 2)
    .fill(C.earInner)
    .circle(0, 0, 9)
    .fill(C.fur)
    .ellipse(7, 2.5, 5, 3.5)
    .fill(C.furLight)
    .roundRect(-2, -4, 12, 5, 2.5)
    .fill(C.mask)
    .circle(4, -1.5, 2)
    .fill(C.eye)
    .circle(4.6, -1.3, 1)
    .fill(C.pupil)
    .ellipse(11.5, 1.5, 1.6, 1.3)
    .fill(C.nose);
  head.addChild(face);

  body.addChild(tail, ...farLegs, torso, ...nearLegs, head);
  return { body, torso, head, tail, legs: [...farLegs, ...nearLegs] };
}
