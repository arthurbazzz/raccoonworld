import { Application, Container, Graphics, Text } from "pixi.js";

import type { PlayerId, PlayerState } from "@/net/protocol";

import type { ChatBalloons } from "./chat";
import { playerDepth, tileDepth } from "./depth";
import type { Interpolator } from "./interpolation";
import { RaccoonView } from "./raccoonArt";
import { fitScale, type SolidTile, type TileMap } from "./tilemap";

// Scenery and UI colors; the raccoon palette lives in raccoonArt.ts.
const COLORS = {
  background: 0x2b2622,
  floorA: 0x8fb07a,
  floorB: 0x86a872,
  blockTop: 0x9b8a74,
  blockFront: 0x6e604f,
  shadow: 0x000000,
  label: 0xffffff,
  youLabel: 0xffe08a,
  balloon: 0xffffff,
  balloonText: 0x2b2622,
};

/** How tall a solid tile looks in the 3/4 view, in pixels. */
const BLOCK_HEIGHT = 14;
/** The nickname sits just above the ears; chat balloons above the nickname. */
const LABEL_BOTTOM = -40;
const BALLOON_BOTTOM = -56;
const BALLOON_MAX_WIDTH = 160;

interface RaccoonSprite {
  body: RaccoonView;
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

  /**
   * Moves raccoons to their interpolated position, turns and animates them
   * (RW-002), and updates balloons.
   */
  private place(now: number): void {
    for (const [id, sprite] of this.raccoons) {
      const pos = this.interpolator.position(id, now);
      if (pos) {
        sprite.body.position.set(pos.x, pos.y);
        sprite.body.zIndex = playerDepth(pos.y);
      }
      const motion = this.interpolator.motion(id, now);
      if (motion) {
        if (motion.facing !== sprite.body.facing) sprite.body.setFacing(motion.facing);
        sprite.body.animate(motion.moving, now);
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

/** A raccoon (RW-002 art) with its nickname above the head. */
function drawRaccoon(nickname: string, isYou: boolean): RaccoonView {
  const view = new RaccoonView({ isYou });
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
  label.position.set(0, LABEL_BOTTOM);
  view.addChild(label); // outside the body, so it never flips
  return view;
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
