"use client";

import type { CupsState } from "@/game/cups/cupsState";
import { cupsAt } from "@/game/cups/timeline";
import { useFrameTime } from "@/game/cups/useFrameTime";

const LABELS = {
  idle: "livre",
  lobby: "esperando jogadores",
  show: "mostrando a noz",
  shuffle: "embaralhando",
  choose: "esperando as escolhas",
  reveal: "copos levantados",
  results: "mostrando o placar",
} as const;

/**
 * What the booth in the Praça is doing, for screen readers and UI tests:
 * the booth itself is drawn on the canvas (RC20). Follows the same timeline
 * as the drawing.
 */
export function BoothStatus({ cups, className }: { cups: CupsState; className?: string }) {
  const active = cups.phase === "playing" && cups.round !== null;
  const now = useFrameTime(active);
  let stage: keyof typeof LABELS = cups.phase === "playing" ? "show" : cups.phase;
  if (active && cups.round) {
    stage =
      cups.reveal?.round === cups.round.round
        ? "reveal"
        : cupsAt(cups.round, now - cups.round.startedAt).stage;
  }
  return (
    <p className={className} aria-label="Barraca dos três copos" data-stage={stage}>
      Barraca dos três copos: {LABELS[stage]}
    </p>
  );
}
