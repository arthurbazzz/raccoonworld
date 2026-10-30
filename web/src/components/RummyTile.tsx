"use client";

import type { CSSProperties, PointerEvent } from "react";

import type { RummyManifest } from "@/game/rummy/sheet";
import { COLORS, tileName, tileOf } from "@/game/rummy/tiles";

import styles from "./RummyPanel.module.css";

/** Drawn tile height in CSS pixels; the width follows the art. */
export const TILE_H = 58;

interface Props {
  id: number;
  manifest: RummyManifest | null;
  draggable?: boolean;
  dragging?: boolean;
  onPointerDown?: (e: PointerEvent<HTMLDivElement>, id: number) => void;
}

/** The art's box for a tile; the joker is cropped to a tile's width. */
function artStyle(id: number, m: RummyManifest): CSSProperties {
  const t = tileOf(id);
  const sample = m.tiles[0][0];
  const box = t.joker ? m.joker : m.tiles[COLORS.indexOf(t.color!)][t.value - 1];
  const scale = TILE_H / box.h;
  const width = sample.w * scale;
  const x = t.joker ? box.x + (box.w - sample.w) / 2 : box.x;
  return {
    width,
    backgroundImage: `url(${m.image})`,
    backgroundSize: `${m.size.w * scale}px ${m.size.h * scale}px`,
    backgroundPosition: `${-x * scale}px ${-box.y * scale}px`,
  };
}

/** One tile, from the art when it loaded, or as a plain colored number. */
export function RummyTile({ id, manifest, draggable, dragging, onPointerDown }: Props) {
  const t = tileOf(id);
  const art = manifest ? artStyle(id, manifest) : null;
  const className = [
    styles.tile,
    art ? styles.art : styles.plain,
    draggable ? styles.draggable : "",
    dragging ? styles.dragging : "",
  ].join(" ");
  return (
    <div
      className={className}
      style={art ?? undefined}
      data-tile={id}
      data-color={t.color ?? "joker"}
      role="img"
      aria-label={tileName(id)}
      onPointerDown={draggable && onPointerDown ? (e) => onPointerDown(e, id) : undefined}
    >
      {!art && (t.joker ? "🦝" : t.value)}
    </div>
  );
}
