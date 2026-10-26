// The raccoon drawn from the sprite sheet (RW-004): public API used by the
// scene and the gallery.
export { frameFor, IDLE_FRAME_MS, WALK_FPS, type FrameRef } from "./frames";
export type { SheetManifest } from "./manifest";
export { RaccoonView, TARGET_HEIGHT } from "./RaccoonSprite";
export { loadRaccoonSheet, MANIFEST_URL, sheetFrom, type RaccoonSheet } from "./sheet";
