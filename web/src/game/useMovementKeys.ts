"use client";

import { useEffect } from "react";

import type { Keys } from "@/net/protocol";

import { isTypingTarget, KeyTracker } from "./keyboard";

/**
 * Listens to WASD on the window while enabled and calls onChange with the
 * held keys whenever they change. Key presses inside text fields are
 * ignored (RN3.1); releases always count, so no key gets stuck when focus
 * moves into the chat while walking.
 */
export function useMovementKeys(enabled: boolean, onChange: (keys: Keys) => void): void {
  useEffect(() => {
    if (!enabled) return;
    const tracker = new KeyTracker();

    const onDown = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey || e.altKey || isTypingTarget(e.target)) return;
      const keys = tracker.down(e.code);
      if (keys) onChange(keys);
    };
    const onUp = (e: KeyboardEvent) => {
      const keys = tracker.up(e.code);
      if (keys) onChange(keys);
    };
    const onBlur = () => {
      const keys = tracker.reset();
      if (keys) onChange(keys);
    };

    window.addEventListener("keydown", onDown);
    window.addEventListener("keyup", onUp);
    window.addEventListener("blur", onBlur);
    return () => {
      window.removeEventListener("keydown", onDown);
      window.removeEventListener("keyup", onUp);
      window.removeEventListener("blur", onBlur);
      onBlur(); // stop walking when the game unmounts
    };
  }, [enabled, onChange]);
}
