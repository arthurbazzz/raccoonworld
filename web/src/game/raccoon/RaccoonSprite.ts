import { Container, Graphics, Sprite } from "pixi.js";

import type { Facing } from "../motion";

import { frameFor } from "./frames";
import type { RaccoonSheet } from "./sheet";

// The raccoon drawn from the AI sprite sheet (RW-004). Same interface the
// scene used with the RW-002 art: setFacing, animate, ring, facing.

/** Height of the raccoon in the room, in pixels (RW-004, RS4). */
export const TARGET_HEIGHT = 46;

const RING_COLOR = 0xffe08a; // same yellow as your nickname

/**
 * A raccoon sprite standing on its feet at (0, 0). Only the sprite flips
 * when facing left, so the ring, the shadow and a nickname added to the
 * view stay readable.
 */
export class RaccoonView extends Container {
  facing: Facing = "down";
  readonly sprite: Sprite;
  /** RW-002 RV3: the ring under your own raccoon, or null for everyone else. */
  readonly ring: Graphics | null;
  private readonly scaleFactor: number;
  private moving = false;
  private t = 0;

  constructor(
    private readonly sheet: RaccoonSheet,
    { isYou = false }: { isYou?: boolean } = {},
  ) {
    super();
    const { manifest } = sheet;
    this.ring = isYou ? this.addChild(drawRing()) : null;
    this.addChild(drawShadow());

    this.sprite = new Sprite(sheet.frames[manifest.rows.front][manifest.idle[0]]);
    this.sprite.anchor.set(
      manifest.anchor.x / manifest.cell.w,
      manifest.anchor.y / manifest.cell.h,
    );
    this.scaleFactor = TARGET_HEIGHT / manifest.artHeight;
    this.sprite.scale.set(this.scaleFactor);
    this.addChild(this.sprite);
  }

  setFacing(facing: Facing): void {
    this.facing = facing;
    this.show();
  }

  /** Shows the walking or idle frame for time `t` (ms). */
  animate(moving: boolean, t: number): void {
    this.moving = moving;
    this.t = t;
    this.show();
  }

  private show(): void {
    const ref = frameFor(this.sheet.manifest, this.facing, this.moving, this.t);
    this.sprite.texture = this.sheet.frames[ref.row][ref.col];
    this.sprite.scale.x = ref.mirrored ? -this.scaleFactor : this.scaleFactor;
  }
}

/** A flat ring on the ground around the feet (RW-002 RV3). */
function drawRing(): Graphics {
  return new Graphics()
    .ellipse(0, 0, 15, 6)
    .fill({ color: RING_COLOR, alpha: 0.18 })
    .stroke({ color: RING_COLOR, width: 2, alpha: 0.9 });
}

function drawShadow(): Graphics {
  return new Graphics().ellipse(0, 0, 12, 4).fill({ color: 0x000000, alpha: 0.22 });
}
