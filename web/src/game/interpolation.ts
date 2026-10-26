import type { PlayerId, ServerMessage } from "@/net/protocol";

import { facingFrom, hasMoved, type Facing, type Motion } from "./motion";

/** Server tick (usecase.TickInterval): snapshots arrive about this often. */
export const TICK_MS = 50;

/**
 * Players are drawn this far in the past (plan §4): two ticks, so there is
 * almost always a later snapshot to blend towards and jitter is hidden.
 */
export const INTERPOLATION_DELAY_MS = 2 * TICK_MS;

/** A gap longer than this between samples means the player was idle. */
const IDLE_GAP_MS = 2 * TICK_MS;

const MAX_SAMPLES = 32;

/** RV2: a raccoon counts as walking if it moved within this window. */
export const MOVING_WINDOW_MS = 150;

/** RV1: who just arrived faces the camera. */
const INITIAL_FACING: Facing = "down";

interface Sample {
  t: number; // arrival time on the client clock, in ms
  x: number;
  y: number;
}

export interface Point {
  x: number;
  y: number;
}

/**
 * Smooths server positions for rendering. Snapshots are timestamped when
 * they arrive and rendered INTERPOLATION_DELAY_MS later, blending the two
 * samples around that moment. There is no extrapolation: past the newest
 * sample a player simply stands still, as the server is authoritative.
 */
export class Interpolator {
  private readonly samples = new Map<PlayerId, Sample[]>();

  /** Feeds a server message received at `now` (performance.now()). */
  ingest(msg: ServerMessage, now: number): void {
    switch (msg.type) {
      case "room.joined":
        this.samples.clear();
        for (const p of msg.data.players) this.samples.set(p.id, [{ t: now, x: p.x, y: p.y }]);
        break;
      case "player.joined":
        this.samples.set(msg.data.id, [{ t: now, x: msg.data.x, y: msg.data.y }]);
        break;
      case "player.left":
        this.samples.delete(msg.data.id);
        break;
      case "room.snapshot":
        for (const p of msg.data.players) this.push(p.id, { t: now, x: p.x, y: p.y });
        break;
    }
  }

  /** Where to draw a player at `now`, or null if unknown. */
  position(id: PlayerId, now: number): Point | null {
    const s = this.samples.get(id);
    if (!s || s.length === 0) return null;
    const t = now - INTERPOLATION_DELAY_MS;

    if (t <= s[0].t) return { x: s[0].x, y: s[0].y };
    const last = s[s.length - 1];
    if (t >= last.t) return { x: last.x, y: last.y };

    let i = s.length - 2;
    while (i > 0 && s[i].t > t) i--;
    const a = s[i];
    const b = s[i + 1];
    const k = (t - a.t) / (b.t - a.t);
    return { x: a.x + (b.x - a.x) * k, y: a.y + (b.y - a.y) * k };
  }

  /**
   * Facing and walking state at `now`, from the same rendered positions as
   * position(), so a raccoon turns exactly when it starts moving on screen.
   */
  motion(id: PlayerId, now: number): Motion | null {
    const s = this.samples.get(id);
    if (!s || s.length === 0) return null;
    const t = now - INTERPOLATION_DELAY_MS;

    // Facing: the latest segment already started at render time that moved.
    let facing = INITIAL_FACING;
    for (let i = s.length - 2; i >= 0; i--) {
      if (s[i].t >= t) continue;
      const dx = s[i + 1].x - s[i].x;
      const dy = s[i + 1].y - s[i].y;
      if (hasMoved(dx, dy)) {
        facing = facingFrom(dx, dy, INITIAL_FACING);
        break;
      }
    }

    const a = this.position(id, now)!;
    const b = this.position(id, now - MOVING_WINDOW_MS)!;
    return { facing, moving: hasMoved(a.x - b.x, a.y - b.y) };
  }

  /** Number of buffered samples for a player (for tests). */
  historySize(id: PlayerId): number {
    return this.samples.get(id)?.length ?? 0;
  }

  private push(id: PlayerId, sample: Sample): void {
    const s = this.samples.get(id);
    if (!s) return; // unknown player: ignore, like the room state does
    const last = s[s.length - 1];
    if (sample.t - last.t > IDLE_GAP_MS) {
      // Coming out of idle: hold still until one tick before this move, so
      // the raccoon walks from where it stood instead of gliding slowly.
      s.push({ t: sample.t - TICK_MS, x: last.x, y: last.y });
    }
    s.push(sample);
    if (s.length > MAX_SAMPLES) s.splice(0, s.length - MAX_SAMPLES);
  }
}
