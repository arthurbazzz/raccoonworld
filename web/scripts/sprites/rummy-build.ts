// `npm run sprites`, second step: finds the pieces of the Rummy das Nozes
// sheet and writes it, as PNG, with a manifest of where each piece is
// (RW-006, Q6). Fails loudly if the art does not have the agreed layout.
//
// Usage: node scripts/sprites/rummy-build.ts [--src art/rummy.webp] [--out public/sprites]

import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";

import sharp from "sharp";

import { rummyManifest } from "./rummy.ts";
import { SpriteSheetError } from "./slice.ts";

function arg(name: string, fallback: string): string {
  const i = process.argv.indexOf(name);
  return i >= 0 && process.argv[i + 1] ? process.argv[i + 1] : fallback;
}

async function main(): Promise<void> {
  const src = arg("--src", "art/rummy.webp");
  const outDir = arg("--out", "public/sprites");

  const { data, info } = await sharp(src).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
  const manifest = rummyManifest(
    { width: info.width, height: info.height, data: new Uint8Array(data) },
    "/sprites/rummy.png",
  );

  await mkdir(outDir, { recursive: true });
  await sharp(src).png().toFile(path.join(outDir, "rummy.png"));
  await writeFile(path.join(outDir, "rummy.json"), JSON.stringify(manifest, null, 2) + "\n");

  const widths = manifest.tiles.flat().map((t) => t.w);
  console.log(
    `sprites: 4×13 peças, coringa, verso e mesa de ${src} → ${outDir}/rummy.png ` +
      `(peças de ${Math.min(...widths)}–${Math.max(...widths)}px de largura)`,
  );
}

main().catch((err: unknown) => {
  if (err instanceof SpriteSheetError) {
    console.error(`sprites: arte do rummy fora do formato (RW-006, Q6): ${err.message}`);
  } else {
    console.error("sprites: falhou", err);
  }
  process.exit(1);
});
