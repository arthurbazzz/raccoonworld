import { Container, Graphics } from "pixi.js";

import { outlineOf, type RaccoonPalette } from "./palette";
import { OUTLINE_WIDTH, pointedEar, ringedTail, shadedCircle, shadedEllipse } from "./shapes";

// The poses of the raccoon, built in layers (RW-003, RA2) in the style of
// the reference sprite sheet: big head (about half of the height), small
// body, short limbs. Everything grows upward from the feet at (0, 0).

/** The parts of a pose that are animated or dressed. */
export interface RaccoonParts {
  /** Everything that flips when facing left. */
  body: Container;
  torso: Graphics;
  head: Container;
  /** The top layer of the head: mask, eyes, nose, mouth. */
  face: Graphics;
  tail: Graphics;
  legLayer: Container;
  legs: Graphics[];
  /** Which pair each leg belongs to (0 or 1): pairs lift in turn when walking. */
  legPairs: (0 | 1)[];
  armLayer: Container;
  arms: Graphics[];
}

/** Where the head sits; the breath animation scales this distance. */
const HEAD_Y = -27;
const HEAD_R = 13;

function limb(x: number, y: number, rx: number, ry: number, p: RaccoonPalette): Graphics {
  return shadedEllipse(new Graphics(), x, y, rx, ry, p.fur, outlineOf(p));
}

/** A small, soft tuft of fur on top of the head, like the reference. */
function tuft(g: Graphics, y: number, p: RaccoonPalette, outline: number): Graphics {
  return g
    .moveTo(-2.4, y + 1.2)
    .quadraticCurveTo(-1.6, y - 1.6, -0.4, y - 1.2)
    .quadraticCurveTo(0.6, y - 2.4, 2.4, y + 1.2)
    .closePath()
    .fill(p.fur)
    .stroke({ color: outline, width: OUTLINE_WIDTH * 0.7, join: "round" });
}

/** Front pose (facing down): face to the camera, tail peeking on the right. */
export function buildFront(p: RaccoonPalette): RaccoonParts {
  const outline = outlineOf(p);
  const body = new Container();

  const tail = ringedTail(new Graphics(), 4.2, 8, p, outline);
  tail.position.set(6, -6);
  tail.rotation = 0.6;

  const legLayer = new Container();
  const legs = [limb(-4.5, -2.6, 3.4, 3, p), limb(4.5, -2.6, 3.4, 3, p)];
  legLayer.addChild(...legs);

  const torso = shadedEllipse(new Graphics(), 0, -11, 8.5, 8.5, p.fur, outline);
  shadedEllipse(torso, 0, -9.6, 5.2, 6, p.furLight, null); // belly

  const armLayer = new Container();
  const arms = [limb(-7.8, -12, 2.6, 4, p), limb(7.8, -12, 2.6, 4, p)];
  armLayer.addChild(...arms);

  const head = new Container();
  head.position.set(0, HEAD_Y);
  const skull = new Graphics();
  pointedEar(skull, -7.5, -8, 9, 10, p, outline, -2.5);
  pointedEar(skull, 7.5, -8, 9, 10, p, outline, 2.5);
  shadedCircle(skull, 0, 0, HEAD_R, p.fur, outline);
  tuft(skull, -HEAD_R + 0.5, p, outline);
  const face = drawFrontFace(p, outline);
  head.addChild(skull, face);

  body.addChild(tail, legLayer, torso, armLayer, head);
  return { body, torso, head, face, tail, legLayer, legs, legPairs: [0, 1], armLayer, arms };
}

/**
 * The face as in the reference: gray fur with a cream rim hugging the top of
 * a big bandit mask, a cream muzzle, big black eyes with a shine, a nose and
 * a small "w" mouth.
 */
function drawFrontFace(p: RaccoonPalette, outline: number): Graphics {
  const g = new Graphics();
  // Cream rim: the mask shapes, a bit bigger and higher, drawn first.
  for (const side of [-1, 1]) {
    g.ellipse(side * 5.2, 0.2, 6.2, 4.8).fill(p.furLight);
  }
  g.ellipse(0, 6.2, 5.6, 4).fill(p.furLight); // muzzle
  // The bandit mask: two big drops around the eyes, swelling down and out,
  // joined at the bridge of the nose.
  for (const side of [-1, 1]) {
    g.ellipse(side * 5.1, 1.4, 5.1, 3.9).fill(p.mask);
    g.ellipse(side * 7.4, 3, 3, 2.6).fill(p.mask);
  }
  g.rect(-1.6, 0.6, 3.2, 2.2).fill(p.mask);
  // Eyes: big and black inside the mask, with a large and a tiny shine.
  for (const side of [-1, 1]) {
    const x = side * 4.9;
    g.circle(x, 1.2, 2.5).fill(p.nose);
    g.circle(x - 0.8, 0.3, 0.95).fill(p.eye);
    g.circle(x + 0.7, 2, 0.4).fill(p.eye);
  }
  g.ellipse(0, 4.7, 1.7, 1.2).fill(p.nose);
  return g
    .moveTo(-1.7, 6.3)
    .quadraticCurveTo(-0.85, 7.4, 0, 6.5)
    .quadraticCurveTo(0.85, 7.4, 1.7, 6.3)
    .stroke({ color: outline, width: OUTLINE_WIDTH * 0.7, cap: "round", join: "round" });
}
