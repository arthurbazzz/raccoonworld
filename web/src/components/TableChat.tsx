"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";

import { MAX_MESSAGE_LENGTH, prepareMessage } from "@/game/chat";
import { isTypingTarget } from "@/game/keyboard";
import type { TableChatLine } from "@/game/rummy/rummyState";
import type { PlayerId } from "@/net/protocol";

import styles from "./RummyTable.module.css";

interface Props {
  open: boolean;
  onToggle: () => void;
  lines: TableChatLine[];
  name: (id: PlayerId) => string;
  you: PlayerId;
  onSend: (text: string) => void;
}

/**
 * The table chat (RN37): only the players of this match read it. Enter
 * anywhere focuses the field, as in the Praça chat.
 */
export function TableChat({ open, onToggle, lines, name, you, onSend }: Props) {
  const [text, setText] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLOListElement>(null);
  const tooLong = [...text].length > MAX_MESSAGE_LENGTH;

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Enter" && !isTypingTarget(e.target)) {
        e.preventDefault();
        if (!open) onToggle();
        requestAnimationFrame(() => inputRef.current?.focus());
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onToggle]);

  useEffect(() => {
    listRef.current?.scrollTo({ top: listRef.current.scrollHeight });
  }, [lines.length, open]);

  const submit = (e: FormEvent) => {
    e.preventDefault();
    const msg = prepareMessage(text);
    if (msg === null || tooLong) return;
    onSend(msg);
    setText("");
  };

  return (
    <aside className={`${styles.chat} ${open ? styles.chatOpen : ""}`} aria-label="Chat da mesa">
      <button type="button" className={styles.chatToggle} onClick={onToggle} aria-expanded={open}>
        💬 Chat da mesa{!open && lines.length > 0 ? ` (${lines.length})` : ""}
      </button>
      {open && (
        <>
          <ol className={styles.chatLines} ref={listRef}>
            {lines.length === 0 && (
              <li className={styles.chatEmpty}>Só quem está nesta mesa lê.</li>
            )}
            {lines.map((l, i) => (
              <li key={i} className={l.from === you ? styles.mine : undefined}>
                <strong>{l.from === you ? "Você" : name(l.from)}:</strong> {l.text}
              </li>
            ))}
          </ol>
          <form className={styles.chatForm} onSubmit={submit}>
            <input
              ref={inputRef}
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Escape") {
                  e.stopPropagation();
                  inputRef.current?.blur();
                }
              }}
              placeholder="Mensagem para a mesa"
              aria-label="Mensagem para a mesa"
              aria-invalid={tooLong}
            />
            <button type="submit" disabled={tooLong || text.trim() === ""}>
              Enviar
            </button>
          </form>
        </>
      )}
    </aside>
  );
}
