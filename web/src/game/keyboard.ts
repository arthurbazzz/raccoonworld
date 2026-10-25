import type { Keys } from "@/net/protocol";

// RN3: W A S D by physical position (KeyboardEvent.code), so the same keys
// work on any keyboard layout.
const KEY_TO_DIRECTION: Record<string, keyof Keys> = {
  KeyW: "up",
  KeyA: "left",
  KeyS: "down",
  KeyD: "right",
};

const NO_KEYS: Keys = { up: false, down: false, left: false, right: false };

/**
 * Tracks which movement keys are held. Every method returns the new key
 * set when it changed, or null when it did not, so callers send
 * player.input only on change (plan §4).
 */
export class KeyTracker {
  private keys: Keys = { ...NO_KEYS };

  down(code: string): Keys | null {
    return this.set(code, true);
  }

  up(code: string): Keys | null {
    return this.set(code, false);
  }

  /** Releases every key, e.g. when the window loses focus. */
  reset(): Keys | null {
    if (!this.keys.up && !this.keys.down && !this.keys.left && !this.keys.right) return null;
    this.keys = { ...NO_KEYS };
    return { ...this.keys };
  }

  private set(code: string, held: boolean): Keys | null {
    const dir = KEY_TO_DIRECTION[code];
    if (!dir || this.keys[dir] === held) return null;
    this.keys = { ...this.keys, [dir]: held };
    return { ...this.keys };
  }
}

/** RN3.1: keys typed into a text field are text, not movement. */
export function isTypingTarget(target: EventTarget | null): boolean {
  if (!target) return false;
  const el = target as { tagName?: string; isContentEditable?: boolean };
  return (
    el.tagName === "INPUT" ||
    el.tagName === "TEXTAREA" ||
    el.tagName === "SELECT" ||
    el.isContentEditable === true
  );
}
