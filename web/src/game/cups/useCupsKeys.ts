"use client";

import { useEffect } from "react";

import { isTypingTarget } from "../keyboard";

interface Handlers {
  /** E near the booth (RC2). */
  canJoin: boolean;
  /** Esc while at the booth (RC13). */
  canLeave: boolean;
  /** 1, 2, 3 while the cups wait for our pick (RC8); asked on each key. */
  canPick: () => boolean;
  onJoin: () => void;
  onLeave: () => void;
  onPick: (slot: number) => void;
}

const PICK_KEYS: Record<string, number> = {
  Digit1: 0,
  Digit2: 1,
  Digit3: 2,
  Numpad1: 0,
  Numpad2: 1,
  Numpad3: 2,
};

/**
 * The three cups keys. Like WASD, they are ignored while typing in the chat
 * (RN3.1), and held keys do not repeat.
 */
export function useCupsKeys(h: Handlers): void {
  const { canJoin, canLeave, canPick, onJoin, onLeave, onPick } = h;
  useEffect(() => {
    const onDown = (e: KeyboardEvent) => {
      if (e.repeat || e.ctrlKey || e.metaKey || e.altKey || isTypingTarget(e.target)) return;
      if (e.code === "KeyE" && canJoin) {
        onJoin();
      } else if (e.code === "Escape" && canLeave) {
        onLeave();
      } else if (e.code in PICK_KEYS && canPick()) {
        onPick(PICK_KEYS[e.code]);
      }
    };
    window.addEventListener("keydown", onDown);
    return () => window.removeEventListener("keydown", onDown);
  }, [canJoin, canLeave, canPick, onJoin, onLeave, onPick]);
}
