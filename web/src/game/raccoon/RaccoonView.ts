import { Container, Graphics } from "pixi.js";

import type { Facing } from "../motion";
import { DEFAULT_PALETTE, type RaccoonPalette } from "./palette";
import { buildFront, type RaccoonParts } from "./rig";

export type { RaccoonParts } from "./rig";

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
  ring: 0xffe08a, // same yellow as your nickname
};

export type Pose = "front" | "back" | "side";

const C = RACCOON_COLORS;

/** Walking cadence (RV2): one paw pair lifts per step. */
export const STEPS_PER_SECOND = 6;
/** Idle breathing cycle (RV2). */
export const BREATH_PERIOD_MS = 2000;

const BOB_PX = 2;
const LEG_LIFT_PX = 2;
const TAIL_SWAY_RAD = 0.15;
const BREATH_AMOUNT = 0.03;

/** How the animatable parts are offset from their rest pose at one moment. */
export interface AnimationFrame {
  /** Vertical body offset in px (negative is up). */
  bob: number;
  /** How high each paw pair is lifted, in px. */
  legLift: [number, number];
  /** Tail rotation added to its rest angle, in radians. */
  tailSway: number;
  /** Torso vertical scale. */
  breath: number;
}

/**
 * The animation at time `t` (ms), computed from the clock alone so no state
 * is kept per frame. Walking: each step is half a sine period, lifting one
 * paw pair while the body bobs. Idle: a slow breath and nothing else.
 */
export function animationFrame(moving: boolean, t: number): AnimationFrame {
  if (!moving) {
    const breath = BREATH_AMOUNT * (1 - Math.cos((2 * Math.PI * t) / BREATH_PERIOD_MS)) * 0.5;
    return { bob: 0, legLift: [0, 0], tailSway: 0, breath: 1 + breath };
  }
  const s = Math.sin((Math.PI * t * STEPS_PER_SECOND) / 1000);
  return {
    bob: -Math.abs(s) * BOB_PX,
    legLift: [Math.max(0, s) * LEG_LIFT_PX, Math.max(0, -s) * LEG_LIFT_PX],
    tailSway: s * TAIL_SWAY_RAD,
    breath: 1,
  };
}

/** Rest values of the parts, so animate() can offset from them. */
interface Rest {
  legY: number[];
  headY: number;
  tailRotation: number;
}

function restOf(parts: RaccoonParts): Rest {
  return {
    legY: parts.legs.map((l) => l.y),
    headY: parts.head.y,
    tailRotation: parts.tail.rotation,
  };
}

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
  /** RV3: the ring under your own raccoon, or null for everyone else. */
  readonly ring: Graphics | null;
  readonly palette: RaccoonPalette;
  private rest: Rest;

  constructor({
    isYou = false,
    palette = DEFAULT_PALETTE,
  }: { isYou?: boolean; palette?: RaccoonPalette } = {}) {
    super();
    this.palette = palette;
    this.ring = isYou ? this.addChild(drawRing()) : null;
    this.addChild(drawShadow());
    this.parts = buildPose(this.pose, this.palette);
    this.rest = restOf(this.parts);
    this.addChild(this.parts.body);
  }

  setFacing(facing: Facing): void {
    this.facing = facing;
    const pose = poseFor(facing);
    if (pose !== this.pose) {
      const index = this.getChildIndex(this.parts.body);
      this.parts.body.destroy({ children: true });
      this.parts = buildPose(pose, this.palette);
      this.rest = restOf(this.parts);
      this.addChildAt(this.parts.body, index);
      this.pose = pose;
    }
    this.parts.body.scale.x = facing === "left" ? -1 : 1;
  }

  /** Applies the walking or idle animation at time `t` (ms). */
  animate(moving: boolean, t: number): void {
    const f = animationFrame(moving, t);
    const { parts, rest } = this;
    parts.body.y = f.bob;
    parts.legs.forEach((leg, i) => {
      leg.y = rest.legY[i] - f.legLift[parts.legPairs[i]];
    });
    parts.tail.rotation = rest.tailRotation + f.tailSway;
    // The torso grows upward from the feet; the head rides on top of it.
    parts.torso.scale.y = f.breath;
    parts.head.y = rest.headY * f.breath;
  }
}

/** A flat ring on the ground around the feet, outside the body so it never flips or bobs. */
function drawRing(): Graphics {
  return new Graphics()
    .ellipse(0, 0, 15, 6)
    .fill({ color: C.ring, alpha: 0.18 })
    .stroke({ color: C.ring, width: 2, alpha: 0.9 });
}

function drawShadow(): Graphics {
  return new Graphics().ellipse(0, 0, 11, 4).fill({ color: C.shadow, alpha: 0.22 });
}

function buildPose(pose: Pose, palette: RaccoonPalette): RaccoonParts {
  switch (pose) {
    case "front":
      return buildFront(palette);
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

/**
 * RW-002 back and side poses, kept until RW-003 T6/T7 redraw them. They have
 * no separate leg or arm layers, so those are empty containers.
 */
function legacyParts(p: Omit<RaccoonParts, "legLayer" | "armLayer" | "arms">): RaccoonParts {
  return { ...p, legLayer: new Container(), armLayer: new Container(), arms: [] };
}

function headFace(head: Container): Graphics {
  return head.children[0] as Graphics;
}

function ears(g: Graphics, dx: number, y: number): Graphics {
  for (const x of [-dx, dx]) {
    g.circle(x, y, 4).fill(C.fur);
    g.circle(x, y + 0.5, 2).fill(C.earInner);
  }
  return g;
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
  return legacyParts({ body, torso, head, face: headFace(head), tail, legs, legPairs: [0, 1] });
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
  // Diagonal pairs, like a real walk: far-back with near-front.
  return legacyParts({
    body,
    torso,
    head,
    face,
    tail,
    legs: [...farLegs, ...nearLegs],
    legPairs: [0, 1, 1, 0],
  });
}
