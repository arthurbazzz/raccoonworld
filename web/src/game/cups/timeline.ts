import type { CupsRoundData } from "@/net/protocol";

// Where every cup is at any moment of a round (plan, section 4.1). The
// panel and the booth in the Praça both draw from this, so they always
// agree (RC20). It depends only on the time since the round started: a tab
// that was in the background jumps straight to the right place.

export type CupsStage = "show" | "shuffle" | "choose";

export interface CupPose {
  /** Horizontal place, in slots: 0, 1, 2, or in between while swapping. */
  slot: number;
  /** Swing while swapping: one cup passes in front (+), the other behind (-). */
  arc: number;
  /** How far the cup is raised, 0 (down) to 1 (up), showing what is under it. */
  lift: number;
}

export interface CupsFrame {
  stage: CupsStage;
  /** One pose per cup, by cup identity (its starting slot). */
  cups: CupPose[];
  /** Identity of the cup with the nut. */
  nutCup: number;
}

/** Share of the show stage spent raising, and then lowering, the nut's cup. */
const LIFT_SHARE = 0.2;

const smooth = (p: number) => p * p * (3 - 2 * p);

/** Slot of every cup identity after the first `n` swaps. */
function slotsAfter(round: CupsRoundData, n: number): number[] {
  const slots = [0, 1, 2];
  for (const [a, b] of round.swaps.slice(0, n)) {
    const ia = slots.indexOf(a);
    const ib = slots.indexOf(b);
    slots[ia] = b;
    slots[ib] = a;
  }
  return slots;
}

/** Where the nut ends up: what a player who watched every swap answers. */
export function finalNutSlot(round: CupsRoundData): number {
  return slotsAfter(round, round.swaps.length)[round.nut];
}

/** When the cups stop moving and players can pick, in ms from the start. */
export function shuffleEndMs(round: CupsRoundData): number {
  return round.showMs + round.swaps.length * round.swapMs;
}

/** The cups `t` ms after the round started. */
export function cupsAt(round: CupsRoundData, t: number): CupsFrame {
  const nutCup = round.nut;
  if (t < round.showMs) {
    const p = Math.max(0, t) / round.showMs;
    const lift = Math.min(1, p / LIFT_SHARE, (1 - p) / LIFT_SHARE);
    const cups = [0, 1, 2].map((slot) => ({ slot, arc: 0, lift: slot === nutCup ? lift : 0 }));
    return { stage: "show", cups, nutCup };
  }

  const into = (t - round.showMs) / round.swapMs;
  const done = Math.floor(into);
  if (done >= round.swaps.length) {
    const cups = slotsAfter(round, round.swaps.length).map((slot) => ({ slot, arc: 0, lift: 0 }));
    return { stage: "choose", cups, nutCup };
  }

  const before = slotsAfter(round, done);
  const [a, b] = round.swaps[done];
  const p = smooth(into - done);
  const swing = Math.sin(Math.PI * p);
  const cups = before.map((slot) => {
    if (slot === a) return { slot: a + (b - a) * p, arc: swing, lift: 0 };
    if (slot === b) return { slot: b + (a - b) * p, arc: -swing, lift: 0 };
    return { slot, arc: 0, lift: 0 };
  });
  return { stage: "shuffle", cups, nutCup };
}
