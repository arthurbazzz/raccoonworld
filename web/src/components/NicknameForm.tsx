"use client";

import { useState, type FormEvent } from "react";

import { nicknameError, rejectMessage } from "@/game/nickname";
import type { Status } from "@/net/client";
import type { RejectReason } from "@/net/protocol";

import styles from "./NicknameForm.module.css";

interface Props {
  initialNickname: string;
  joining: boolean;
  status: Status | "idle";
  rejection: RejectReason | null;
  onJoin: (nickname: string) => void;
}

export function NicknameForm({ initialNickname, joining, status, rejection, onJoin }: Props) {
  const [nickname, setNickname] = useState(initialNickname);
  const [touched, setTouched] = useState(false);

  const localError = touched ? nicknameError(nickname) : null;
  const serverError = rejection ? rejectMessage(rejection) : null;
  const offline = joining && status === "reconnecting";

  function submit(e: FormEvent) {
    e.preventDefault();
    setTouched(true);
    if (nicknameError(nickname) === null) onJoin(nickname);
  }

  return (
    <main className={styles.page}>
      <form className={styles.card} onSubmit={submit} noValidate>
        <h1 className={styles.title}>Raccoons World</h1>
        <p className={styles.subtitle}>Escolha um apelido para entrar na Praça.</p>

        <label className={styles.label} htmlFor="nickname">
          Apelido
        </label>
        <input
          id="nickname"
          name="nickname"
          className={styles.input}
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
          maxLength={32}
          autoComplete="off"
          autoFocus
          disabled={joining}
          aria-invalid={localError !== null || serverError !== null}
          aria-describedby="nickname-error"
        />

        <p id="nickname-error" className={styles.error} role="alert">
          {localError ??
            serverError ??
            (offline ? "Sem conexão com o servidor. Tentando de novo…" : "")}
        </p>

        <button className={styles.button} type="submit" disabled={joining}>
          {joining ? "Entrando…" : "Entrar na Praça"}
        </button>
      </form>
    </main>
  );
}
