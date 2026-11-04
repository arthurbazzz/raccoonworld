import { Assets, Container, Graphics, Rectangle, Sprite, Text, Texture } from "pixi.js";

import type { Rect } from "../tilemap";

import type { RummyManifest } from "./sheet";

// The pieces table in the Praça (RW-006, RN25): the round table of the art,
// over the 'M' tiles, and how many rooms are playing on it. Nobody sees the
// tiles of a match from here.

/** How much wider than its tiles the table is drawn. */
const OVERHANG = 40;

export class TableView extends Container {
  private readonly caption: Text;

  private constructor(box: Rect, art: Container, baseY: number) {
    super();
    this.addChild(art);
    this.caption = new Text({
      text: "",
      style: {
        fontFamily: "system-ui, sans-serif",
        fontSize: 11,
        fontWeight: "700",
        fill: 0xffffff,
        padding: 4, // keeps the last letter from being clipped (see scene.ts)
      },
    });
    this.caption.anchor.set(0.5, 1);
    this.caption.position.set(box.x + box.w / 2, art.y - 2);
    this.addChild(this.caption);
    // Depth: the base of the table, like a solid tile (RN6).
    this.zIndex = baseY;
  }

  /** Builds the view; with no art (manifest null), a plain wooden ellipse. */
  static async create(box: Rect, manifest: RummyManifest | null): Promise<TableView> {
    const width = box.w + OVERHANG;
    const baseY = box.y + box.h;
    if (manifest) {
      try {
        const texture = await Assets.load<Texture>({
          src: manifest.image,
          data: { autoGenerateMipmaps: true, scaleMode: "linear" },
        });
        const t = manifest.table;
        const sprite = new Sprite(
          new Texture({ source: texture.source, frame: new Rectangle(t.x, t.y, t.w, t.h) }),
        );
        sprite.scale.set(width / t.w);
        sprite.anchor.set(0.5, 1);
        sprite.position.set(box.x + box.w / 2, baseY + 10);
        const view = new TableView(box, sprite, baseY);
        view.caption.y = sprite.y - sprite.height - 2;
        return view;
      } catch {
        // Fall through to the plain table.
      }
    }
    const g = new Graphics();
    g.ellipse(box.x + box.w / 2, box.y + box.h / 2, width / 2, box.h / 2 + 6).fill(0x8a5a35);
    g.ellipse(box.x + box.w / 2, box.y + box.h / 2 - 4, width / 2 - 4, box.h / 2 + 2).fill(
      0xb07a4a,
    );
    const view = new TableView(box, g, baseY);
    view.caption.y = box.y - 10;
    return view;
  }

  /** How many rooms are playing (RN25). */
  update(tables: number): void {
    this.caption.text =
      tables === 0 ? "" : tables === 1 ? "1 mesa jogando" : `${tables} mesas jogando`;
  }
}
