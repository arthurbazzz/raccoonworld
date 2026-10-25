"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";

import { MAX_MESSAGE_LENGTH, prepareMessage } from "@/game/chat";
import { isTypingTarget } from "@/game/keyboard";

import styles from "./ChatBox.module.css";

interface Props {
  onSend: (text: string) => void;
}

/**
 * Chat field. Enter anywhere in the game focuses it; Enter in the field
 * sends and gives the keyboard back to WASD; Escape leaves without sending.
 */
export function ChatBox({ onSend }: Props) {
  const [text, setText] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const tooLong = [...text].length > MAX_MESSAGE_LENGTH;

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Enter" && !isTypingTarget(e.target)) {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  function submit(e: FormEvent) {
    e.preventDefault();
    const message = prepareMessage(text);
    if (message !== null) {
      onSend(message);
      setText("");
    }
    inputRef.current?.blur();
  }

  return (
    <form className={styles.box} onSubmit={submit}>
      <input
        ref={inputRef}
        className={styles.input}
        aria-label="Mensagem"
        placeholder="Pressione Enter para conversar"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Escape") e.currentTarget.blur();
        }}
        autoComplete="off"
        aria-invalid={tooLong}
      />
      <span className={tooLong ? styles.countOver : styles.count} aria-live="polite">
        {[...text].length}/{MAX_MESSAGE_LENGTH}
      </span>
    </form>
  );
}
