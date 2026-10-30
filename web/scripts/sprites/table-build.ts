// `npm run sprites`, third step: cuts the table pieces (the felt oval, the
// pile and a tile back) out of the table assets sheet (RW-006, Emenda A).
// The buttons of that sheet are drawn in CSS (RummyTable.module.css).
//
// Usage: node scripts/sprites/table-build.ts [--src art/table.webp] [--out public/sprites/table]

import { mkdir } from "node:fs/promises";
import path from "node:path";

import sharp from "sharp";

/** Where each piece is on the 1536x1024 sheet: left, top, width, height. */
const PIECES = {
  felt: [5, 10, 623, 385],
  deck: [895, 800, 190, 190],
  back: [765, 805, 110, 155],
} as const;

function arg(name: string, fallback: string): string {
  const i = process.argv.indexOf(name);
  return i >= 0 && process.argv[i + 1] ? process.argv[i + 1] : fallback;
}

async function main(): Promise<void> {
  const src = arg("--src", "art/table.webp");
  const outDir = arg("--out", "public/sprites/table");
  const { width, height } = await sharp(src).metadata();
  if (width !== 1536 || height !== 1024) {
    throw new Error(`a folha da mesa precisa ter 1536x1024, tem ${width}x${height}`);
  }

  await mkdir(outDir, { recursive: true });
  for (const [name, [left, top, w, h]] of Object.entries(PIECES)) {
    await sharp(src)
      .extract({ left, top, width: w, height: h })
      .png()
      .toFile(path.join(outDir, `${name}.png`));
  }
  console.log(`sprites: mesa, monte e verso de ${src} → ${outDir}`);
}

main().catch((err: unknown) => {
  console.error("sprites: falhou", err);
  process.exit(1);
});
