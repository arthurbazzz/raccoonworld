"use client";

import type { CSSProperties, PointerEvent } from "react";

import type { RummyManifest } from "@/game/rummy/sheet";
import { COLORS, tileName, tileOf } from "@/game/rummy/tiles";

import styles from "./RummyPanel.module.css";

/** Default drawn tile height in CSS pixels; the width follows the art. */
export const TILE_H = 58;

interface Props {
  id: number;
  manifest: RummyManifest | null;
  /** Drawn height in CSS pixels; the width follows the art. */
  height?: number;
  draggable?: boolean;
  dragging?: boolean;
  /** Marked with Shift+click or the selection box (RN35). */
  marked?: boolean;
  /** To blame for an invalid meld or a refused play (RN12a). */
  blamed?: boolean;
  onPointerDown?: (e: PointerEvent<HTMLDivElement>, id: number) => void;
}

/** The art's box for a tile; the joker is cropped to a tile's width. */
function artStyle(id: number, m: RummyManifest, height: number): CSSProperties {
  const t = tileOf(id);
  const sample = m.tiles[0][0];
  const box = t.joker ? m.joker : m.tiles[COLORS.indexOf(t.color!)][t.value - 1];
  const scale = height / box.h;
  const width = sample.w * scale;
  const x = t.joker ? box.x + (box.w - sample.w) / 2 : box.x;
  return {
    width,
    height,
    backgroundImage: `url(${m.image})`,
    backgroundSize: `${m.size.w * scale}px ${m.size.h * scale}px`,
    backgroundPosition: `${-x * scale}px ${-box.y * scale}px`,
  };
}

/** One tile, from the art when it loaded, or as a plain colored number. */
export function RummyTile({
  id,
  manifest,
  height = TILE_H,
  draggable,
  dragging,
  marked,
  blamed,
  onPointerDown,
}: Props) {
  const t = tileOf(id);
  const art = manifest ? artStyle(id, manifest, height) : null;
  const className = [
    styles.tile,
    art ? styles.art : styles.plain,
    draggable ? styles.draggable : "",
    dragging ? styles.dragging : "",
    marked ? styles.marked : "",
    blamed ? styles.blamed : "",
  ].join(" ");
  const plain = { height, width: height * 0.62, fontSize: height * 0.36 };
  return (
    <div
      className={className}
      style={art ?? plain}
      data-tile={id}
      data-color={t.color ?? "joker"}
      data-marked={marked || undefined}
      data-blamed={blamed || undefined}
      role="img"
      aria-label={tileName(id)}
      onPointerDown={onPointerDown ? (e) => onPointerDown(e, id) : undefined}
    >
      {!art && (t.joker ? "🦝" : t.value)}
    </div>
  );
}
