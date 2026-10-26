import { Container, Graphics, Text } from "pixi.js";

import type { Rect } from "../tilemap";

import type { CupsState } from "./cupsState";
import { cupsAt, type CupPose } from "./timeline";

// The three cups booth in the Praça (RW-005, RC1, RC20), drawn in code with
// simple shapes (Q2). Its cups follow the same timeline as the panel, so
// whoever walks by sees the match as it happens.

const COLORS = {
  counterTop: 0xc08a5a,
  counterFront: 0x7a4a2a,
  post: 0x5e3a20,
  awningA: 0xd9483b,
  awningB: 0xf4ede4,
  cup: 0xd9483b,
  cupShade: 0xa8352b,
  cupShine: 0xf6b0a8,
  nut: 0x8b5a2b,
  nutShine: 0xc48a52,
  sign: 0xffe08a,
};

/** Same 3/4 block height as the other solid tiles. */
const BLOCK_HEIGHT = 14;
const AWNING_HEIGHT = 44;
/** Distance between cup centers on the counter, in pixels. */
export const BOOTH_SLOT_GAP = 26;
/** How high a cup rises to show what is under it. */
const LIFT = 14;
/** How far a swapping cup swings toward or away from the viewer. */
const SWING = 5;

/** A cup upside down, mouth on the counter, its base point at (0, 0). */
export function drawCup(scale = 1): Graphics {
  const g = new Graphics();
  const s = scale;
  g.poly([-8 * s, 0, 8 * s, 0, 5.5 * s, -16 * s, -5.5 * s, -16 * s]).fill(COLORS.cup);
  g.poly([3 * s, 0, 8 * s, 0, 5.5 * s, -16 * s, 2 * s, -16 * s]).fill(COLORS.cupShade);
  g.rect(-4.5 * s, -13 * s, 2 * s, 9 * s).fill(COLORS.cupShine);
  g.ellipse(0, 0, 8 * s, 2.5 * s).fill(COLORS.cupShade);
  return g;
}

/** The nut, centered on the counter at (0, 0). */
export function drawNut(scale = 1): Graphics {
  const g = new Graphics();
  g.ellipse(0, -3 * scale, 5 * scale, 4 * scale).fill(COLORS.nut);
  g.ellipse(-1.5 * scale, -4.5 * scale, 1.6 * scale, 1.2 * scale).fill(COLORS.nutShine);
  return g;
}

/** The booth and its cups; call update() every frame. */
export class BoothView extends Container {
  private readonly cups: Graphics[];
  private readonly nut: Graphics;
  private readonly counterX: number;
  private readonly counterY: number;

  constructor(booth: Rect) {
    super();
    const { x, y, w, h } = booth;
    this.zIndex = y + h; // sorted like a solid tile: by its bottom edge

    const stall = new Graphics();
    const top = y - BLOCK_HEIGHT;
    const roof = top - AWNING_HEIGHT;
    // Posts behind the counter, then the counter (top and front faces).
    stall.rect(x + 3, roof, 4, AWNING_HEIGHT).fill(COLORS.post);
    stall.rect(x + w - 7, roof, 4, AWNING_HEIGHT).fill(COLORS.post);
    stall.rect(x, top, w, h).fill(COLORS.counterTop);
    stall.rect(x, y + h - BLOCK_HEIGHT, w, BLOCK_HEIGHT).fill(COLORS.counterFront);
    // Striped awning with a scalloped edge.
    const stripes = 8;
    const sw = (w + 8) / stripes;
    for (let i = 0; i < stripes; i++) {
      const sx = x - 4 + i * sw;
      const color = i % 2 ? COLORS.awningB : COLORS.awningA;
      stall.rect(sx, roof - 10, sw, 12).fill(color);
      stall.circle(sx + sw / 2, roof + 2, sw / 2).fill(color);
    }
    this.addChild(stall);

    const sign = new Text({
      text: "3 copos",
      style: {
        fontFamily: "system-ui, sans-serif",
        fontSize: 9,
        fontWeight: "700",
        fill: COLORS.sign,
        stroke: { color: 0x000000, width: 3 },
      },
      resolution: 2,
    });
    sign.anchor.set(0.5, 1);
    sign.position.set(x + w / 2, roof - 11);
    this.addChild(sign);

    this.counterX = x + w / 2;
    this.counterY = top + h / 2 + 4;
    this.nut = this.addChild(drawNut());
    this.nut.visible = false;
    this.cups = [0, 1, 2].map(() => this.addChild(drawCup()));
    this.place([0, 1, 2].map((slot) => ({ slot, arc: 0, lift: 0 })));
  }

  /** Poses the cups for the booth state at `now` (performance.now). */
  update(state: CupsState, now: number): void {
    const { round, reveal } = state;
    if (!round || state.phase !== "playing") {
      this.nut.visible = false;
      this.place([0, 1, 2].map((slot) => ({ slot, arc: 0, lift: 0 })));
      return;
    }
    if (reveal && reveal.round === round.round) {
      // Cups up: the nut shows where it ended (RC8).
      this.place([0, 1, 2].map((slot) => ({ slot, arc: 0, lift: 1 })));
      this.showNut(reveal.nut);
      return;
    }
    const frame = cupsAt(round, now - round.startedAt);
    this.place(frame.cups);
    const nutCup = frame.cups[frame.nutCup];
    this.nut.visible = nutCup.lift > 0;
    if (this.nut.visible) this.showNut(nutCup.slot);
  }

  private showNut(slot: number): void {
    this.nut.visible = true;
    this.nut.position.set(this.slotX(slot), this.counterY);
  }

  private slotX(slot: number): number {
    return this.counterX + (slot - 1) * BOOTH_SLOT_GAP;
  }

  private place(poses: CupPose[]): void {
    poses.forEach((pose, i) => {
      const cup = this.cups[i];
      cup.position.set(this.slotX(pose.slot), this.counterY + pose.arc * SWING - pose.lift * LIFT);
      cup.zIndex = pose.arc;
    });
    // The cup swinging toward the viewer is drawn over the one going behind.
    const order = [...this.cups].sort((a, b) => a.zIndex - b.zIndex);
    order.forEach((cup) => this.addChild(cup));
  }
}

/** A small cup before a nickname: this raccoon is playing (RC22). */
export function drawPlayingSign(): Graphics {
  const g = drawCup(0.55);
  g.label = "playing-sign";
  return g;
}
