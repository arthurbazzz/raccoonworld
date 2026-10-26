"use client";

import { useEffect, useRef } from "react";

import type { Facing } from "@/game/motion";

const FACINGS: { facing: Facing; label: string }[] = [
  { facing: "down", label: "frente" },
  { facing: "up", label: "costas" },
  { facing: "right", label: "direita" },
  { facing: "left", label: "esquerda" },
];
const STATES = [
  { moving: false, label: "parado" },
  { moving: true, label: "andando" },
];

const ZOOM = 4;
const CELL = 56 * ZOOM; // one zoomed raccoon per column
const LABEL_H = 26;
const ACTUAL_H = 70; // strip with the same raccoons at game size
// One block per state: label, zoomed row, game-size row.
const BLOCK_H = LABEL_H + CELL + ACTUAL_H;
const FLOOR = 0x8fb07a; // the Praça floor, so colors read as in the game

/**
 * Every pose of the raccoon, idle and walking, zoomed and at game size
 * (RW-003, RA8). Used to judge the art; not part of the game.
 */
export function RaccoonGallery() {
  const hostRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;
    let cancelled = false;
    let destroy: (() => void) | null = null;

    (async () => {
      const [{ Application, Text }, { RaccoonView }] = await Promise.all([
        import("pixi.js"),
        import("@/game/raccoon"),
      ]);
      const app = new Application();
      await app.init({
        width: FACINGS.length * CELL,
        height: STATES.length * BLOCK_H,
        background: FLOOR,
        antialias: true,
        resolution: Math.min(window.devicePixelRatio || 1, 2),
        autoDensity: true,
      });
      if (cancelled) {
        app.destroy(true);
        return;
      }
      host.appendChild(app.canvas);
      destroy = () => app.destroy(true, { children: true });

      const label = (text: string, x: number, y: number) => {
        const t = new Text({
          text,
          style: { fontFamily: "system-ui", fontSize: 13, fill: 0xffffff },
        });
        t.anchor.set(0.5, 0);
        t.position.set(x, y);
        app.stage.addChild(t);
      };

      const views: { view: InstanceType<typeof RaccoonView>; moving: boolean }[] = [];
      STATES.forEach(({ moving, label: stateLabel }, row) => {
        const top = row * BLOCK_H;
        FACINGS.forEach(({ facing, label: facingLabel }, col) => {
          const cx = col * CELL + CELL / 2;
          label(`${facingLabel} · ${stateLabel}`, cx, top + 6);

          const zoomed = new RaccoonView();
          zoomed.setFacing(facing);
          zoomed.scale.set(ZOOM);
          zoomed.position.set(cx, top + LABEL_H + CELL - 12);
          app.stage.addChild(zoomed);

          // Game size below; the first one has the ring, to judge it at 1x too.
          const actual = new RaccoonView({ isYou: col === 0 });
          actual.setFacing(facing);
          actual.position.set(cx, top + BLOCK_H - 14);
          app.stage.addChild(actual);

          views.push({ view: zoomed, moving }, { view: actual, moving });
        });
      });

      app.ticker.add(() => {
        const now = performance.now();
        for (const { view, moving } of views) view.animate(moving, now);
      });
    })();

    return () => {
      cancelled = true;
      destroy?.();
    };
  }, []);

  return (
    <main style={{ padding: 16, overflowX: "auto" }}>
      <h1 style={{ margin: "0 0 4px" }}>Galeria do guaxinim</h1>
      <p style={{ margin: "0 0 12px", opacity: 0.7 }}>
        Só em desenvolvimento (RW-003). Todas as poses, paradas e andando, ampliadas e no tamanho do
        jogo.
      </p>
      <div ref={hostRef} />
    </main>
  );
}
