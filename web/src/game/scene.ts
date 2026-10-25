import { Application, Container, Graphics, Text } from "pixi.js";

import type { PlayerId, PlayerState } from "@/net/protocol";

import type { ChatBalloons } from "./chat";
import { playerDepth, tileDepth } from "./depth";
import type { Interpolator } from "./interpolation";
import { fitScale, type SolidTile, type TileMap } from "./tilemap";

// Placeholder art (plan §4): flat shapes until the real sprites arrive.
const COLORS = {
  background: 0x2b2622,
  floorA: 0x8fb07a,
  floorB: 0x86a872,
  blockTop: 0x9b8a74,
  blockFront: 0x6e604f,
  shadow: 0x000000,
  fur: 0x8a8a8f,
  furLight: 0xd9d6cf,
  mask: 0x2e2e33,
  label: 0xffffff,
  youLabel: 0xffe08a,
  balloon: 0xffffff,
  balloonText: 0x2b2622,
};

/** How tall a solid tile looks in the 3/4 view, in pixels. */
const BLOCK_HEIGHT = 14;
/** Chat balloons sit above the nickname label. */
const BALLOON_BOTTOM = -54;
const BALLOON_MAX_WIDTH = 160;

interface RaccoonSprite {
  body: Container;
  balloon: Container | null;
  balloonText: string | null;
}

/**
 * Renders one room with PixiJS. It only draws: who is in the room comes
 * from sync(), positions from the interpolator and chat from the balloons,
 * all fed by server messages, never by local input.
 */
export class RoomScene {
  private readonly world = new Container();
  private readonly floor = new Container();
  private readonly entities = new Container({ sortableChildren: true });
  // Balloons are drawn above everything so scenery never hides a message.
  private readonly overlay = new Container();
  private readonly raccoons = new Map<PlayerId, RaccoonSprite>();
  private readonly onResize = () => this.layout();
  private readonly onFrame = () => this.place(performance.now());

  private constructor(
    private readonly app: Application,
    private readonly map: TileMap,
    private readonly interpolator: Interpolator,
    private readonly balloons: ChatBalloons,
  ) {
    this.world.addChild(this.floor, this.entities, this.overlay);
    app.stage.addChild(this.world);
    this.drawFloor();
    for (const tile of map.solidTiles()) this.entities.addChild(drawBlock(tile));
    app.renderer.on("resize", this.onResize);
    app.ticker.add(this.onFrame);
    this.layout();
  }

  static async create(
    parent: HTMLElement,
    map: TileMap,
    interpolator: Interpolator,
    balloons: ChatBalloons,
  ): Promise<RoomScene> {
    const app = new Application();
    await app.init({
      resizeTo: parent,
      background: COLORS.background,
      antialias: true,
      autoDensity: true,
      resolution: Math.min(window.devicePixelRatio || 1, 2),
    });
    parent.appendChild(app.canvas);
    return new RoomScene(app, map, interpolator, balloons);
  }

  /**
   * Mirrors who is in the room: adds and removes raccoons. Positions are
   * set every frame from the interpolator; the state position is only the
   * fallback for a player the interpolator has not seen.
   */
  sync(players: Record<PlayerId, PlayerState>, you: PlayerId | null): void {
    for (const [id, sprite] of this.raccoons) {
      if (!(id in players)) {
        sprite.body.destroy({ children: true });
        sprite.balloon?.destroy({ children: true });
        this.raccoons.delete(id);
      }
    }
    for (const p of Object.values(players)) {
      if (this.raccoons.has(p.id)) continue;
      const body = drawRaccoon(p.nickname, p.id === you);
      body.position.set(p.x, p.y);
      body.zIndex = playerDepth(p.y);
      this.entities.addChild(body);
      this.raccoons.set(p.id, { body, balloon: null, balloonText: null });
    }
    this.place(performance.now());
  }

  destroy(): void {
    this.app.ticker.remove(this.onFrame);
    this.app.renderer.off("resize", this.onResize);
    this.app.destroy({ removeView: true }, { children: true });
  }

  /** Moves raccoons to their interpolated position and updates balloons. */
  private place(now: number): void {
    for (const [id, sprite] of this.raccoons) {
      const pos = this.interpolator.position(id, now);
      if (pos) {
        sprite.body.position.set(pos.x, pos.y);
        sprite.body.zIndex = playerDepth(pos.y);
      }
      this.updateBalloon(sprite, this.balloons.textFor(id, now));
      sprite.balloon?.position.copyFrom(sprite.body.position);
    }
  }

  private updateBalloon(sprite: RaccoonSprite, text: string | null): void {
    if (text === sprite.balloonText) return;
    sprite.balloon?.destroy({ children: true });
    sprite.balloon = text === null ? null : drawBalloon(text);
    sprite.balloonText = text;
    if (sprite.balloon) this.overlay.addChild(sprite.balloon);
  }

  /** Scales the room to fit the canvas and centers it. */
  private layout(): void {
    const { width, height } = this.app.screen;
    const scale = fitScale(width, height, this.map.width, this.map.height);
    this.world.scale.set(scale);
    this.world.position.set(
      (width - this.map.width * scale) / 2,
      (height - this.map.height * scale) / 2,
    );
  }

  private drawFloor(): void {
    const g = new Graphics();
    const { tileSize: s } = this.map;
    for (let row = 0; row < this.map.rows; row++) {
      for (let col = 0; col < this.map.cols; col++) {
        g.rect(col * s, row * s, s, s).fill((col + row) % 2 ? COLORS.floorA : COLORS.floorB);
      }
    }
    this.floor.addChild(g);
  }
}

/** A solid tile as a short block: lighter top face, darker front face. */
function drawBlock(tile: SolidTile): Graphics {
  const { x, y, size } = tile;
  const g = new Graphics();
  g.rect(x, y + size - BLOCK_HEIGHT, size, BLOCK_HEIGHT).fill(COLORS.blockFront);
  g.rect(x, y - BLOCK_HEIGHT, size, size).fill(COLORS.blockTop);
  g.zIndex = tileDepth(tile);
  return g;
}

/** Placeholder raccoon, drawn upward from its feet at (0, 0). */
function drawRaccoon(nickname: string, isYou: boolean): Container {
  const c = new Container();
  const g = new Graphics();
  g.ellipse(0, 0, 11, 4).fill({ color: COLORS.shadow, alpha: 0.25 });
  // Striped tail.
  g.ellipse(11, -8, 5, 9).fill(COLORS.fur);
  g.rect(7, -10, 9, 3).fill(COLORS.mask);
  // Body and head.
  g.ellipse(0, -10, 10, 10).fill(COLORS.fur);
  g.circle(0, -24, 9).fill(COLORS.fur);
  g.circle(-6, -32, 3).fill(COLORS.fur);
  g.circle(6, -32, 3).fill(COLORS.fur);
  g.ellipse(0, -21, 5, 3).fill(COLORS.furLight);
  // Bandit mask and eyes.
  g.rect(-8, -27, 16, 5).fill(COLORS.mask);
  g.circle(-3.5, -25, 1.5).fill(COLORS.label);
  g.circle(3.5, -25, 1.5).fill(COLORS.label);
  c.addChild(g);

  const label = new Text({
    text: nickname,
    style: {
      fontFamily: "system-ui, sans-serif",
      fontSize: 11,
      fontWeight: isYou ? "700" : "500",
      fill: isYou ? COLORS.youLabel : COLORS.label,
      stroke: { color: COLORS.shadow, width: 3 },
    },
    resolution: 2,
  });
  label.anchor.set(0.5, 1);
  label.position.set(0, -38);
  c.addChild(label);
  return c;
}

/** A speech balloon whose tail points down at (0, BALLOON_BOTTOM). */
function drawBalloon(text: string): Container {
  const c = new Container();
  const label = new Text({
    text,
    style: {
      fontFamily: "system-ui, sans-serif",
      fontSize: 11,
      fill: COLORS.balloonText,
      wordWrap: true,
      breakWords: true,
      wordWrapWidth: BALLOON_MAX_WIDTH,
      align: "center",
    },
    resolution: 2,
  });
  const padX = 7;
  const padY = 4;
  const w = label.width + padX * 2;
  const h = label.height + padY * 2;
  const top = BALLOON_BOTTOM - h;

  const bg = new Graphics();
  bg.roundRect(-w / 2, top, w, h, 7).fill(COLORS.balloon);
  bg.poly([-5, BALLOON_BOTTOM, 5, BALLOON_BOTTOM, 0, BALLOON_BOTTOM + 6]).fill(COLORS.balloon);
  label.position.set(-label.width / 2, top + padY);
  c.addChild(bg, label);
  return c;
}
