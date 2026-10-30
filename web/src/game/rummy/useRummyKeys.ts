"use client";

import { useEffect } from "react";

import { isTypingTarget } from "../keyboard";

interface Handlers {
  /** E near the table, when not at a minigame (RN4). */
  canOpen: boolean;
  /** Esc with the room list open or at the table (RN20). */
  canClose: boolean;
  onOpen: () => void;
  onClose: () => void;
}

/** The pieces table keys; ignored while typing in the chat, like WASD. */
export function useRummyKeys({ canOpen, canClose, onOpen, onClose }: Handlers): void {
  useEffect(() => {
    const onDown = (e: KeyboardEvent) => {
      if (e.repeat || e.ctrlKey || e.metaKey || e.altKey || isTypingTarget(e.target)) return;
      if (e.code === "KeyE" && canOpen) onOpen();
      else if (e.code === "Escape" && canClose) onClose();
    };
    window.addEventListener("keydown", onDown);
    return () => window.removeEventListener("keydown", onDown);
  }, [canOpen, canClose, onOpen, onClose]);
}
