import { Assets, Rectangle, Texture, type TextureSource } from "pixi.js";

import type { SheetManifest } from "./manifest";

/** The raccoon art ready to draw: the manifest and one texture per cell. */
export interface RaccoonSheet {
  manifest: SheetManifest;
  /** frames[row][col], following the manifest grid. */
  frames: Texture[][];
}

/** Where `npm run sprites` writes the manifest (web/public/sprites). */
export const MANIFEST_URL = "/sprites/raccoon.json";

/** Cuts the aligned sheet into one texture per cell. */
export function sheetFrom(manifest: SheetManifest, source: TextureSource): RaccoonSheet {
  const { w, h } = manifest.cell;
  const rows = Math.round(source.height / h);
  const cols = Math.round(source.width / w);
  const frames = Array.from({ length: rows }, (_, r) =>
    Array.from(
      { length: cols },
      (_, c) => new Texture({ source, frame: new Rectangle(c * w, r * h, w, h) }),
    ),
  );
  return { manifest, frames };
}

/**
 * Loads the raccoon sheet once (scene and gallery share it). The art is
 * drawn at ~0.4x, so the texture gets mipmaps and linear filtering to stay
 * smooth instead of jagged.
 */
export async function loadRaccoonSheet(url: string = MANIFEST_URL): Promise<RaccoonSheet> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`raccoon sprites: HTTP ${res.status} for ${url}`);
  const manifest = (await res.json()) as SheetManifest;
  const texture = await Assets.load<Texture>({
    src: manifest.image,
    data: { autoGenerateMipmaps: true, scaleMode: "linear" },
  });
  return sheetFrom(manifest, texture.source);
}
