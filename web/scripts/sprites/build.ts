// `npm run sprites`: turns the AI-made raccoon sheet into an aligned sheet
// plus a manifest for the game (RW-004, RS8). Runs before dev, build and the
// browser tests, and fails loudly if the art does not have the agreed layout.
//
// Usage: node scripts/sprites/build.ts [--src art/raccoon.png] [--out public/sprites]

import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";

import sharp from "sharp";

import { layoutSheet, manifestFor } from "./layout.ts";
import { findFrames, SpriteSheetError } from "./slice.ts";

const ROWS = 3; // front, back, side
const COLS = 6; // 2 idle + 4 walk

function arg(name: string, fallback: string): string {
  const i = process.argv.indexOf(name);
  return i >= 0 && process.argv[i + 1] ? process.argv[i + 1] : fallback;
}

async function main(): Promise<void> {
  const src = arg("--src", "art/raccoon.png");
  const outDir = arg("--out", "public/sprites");

  const { data, info } = await sharp(src).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
  const frames = findFrames(
    { width: info.width, height: info.height, data: new Uint8Array(data) },
    { rows: ROWS, cols: COLS },
  );
  const layout = layoutSheet(frames);

  const pieces = await Promise.all(
    layout.placements.map(async (p) => ({
      input: await sharp(src)
        .extract({
          left: p.src.x0,
          top: p.src.top,
          width: p.src.x1 - p.src.x0 + 1,
          height: p.src.bottom - p.src.top + 1,
        })
        .png()
        .toBuffer(),
      left: p.x,
      top: p.y,
    })),
  );

  await mkdir(outDir, { recursive: true });
  await sharp({
    create: {
      width: layout.size.w,
      height: layout.size.h,
      channels: 4,
      background: { r: 0, g: 0, b: 0, alpha: 0 },
    },
  })
    .composite(pieces)
    .png()
    .toFile(path.join(outDir, "raccoon.png"));
  const manifest = manifestFor(layout, "/sprites/raccoon.png");
  await writeFile(path.join(outDir, "raccoon.json"), JSON.stringify(manifest, null, 2) + "\n");

  console.log(
    `sprites: ${ROWS}×${COLS} quadros de ${src} → ${outDir}/raccoon.png ` +
      `(${layout.size.w}×${layout.size.h}, célula ${layout.cell.w}×${layout.cell.h}, ` +
      `altura da arte ${layout.artHeight}px)`,
  );
}

main().catch((err: unknown) => {
  if (err instanceof SpriteSheetError) {
    console.error(`sprites: imagem fora do formato (RW-004, RS2): ${err.message}`);
  } else {
    console.error("sprites: falhou", err);
  }
  process.exit(1);
});
