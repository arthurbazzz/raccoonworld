"use client";

import { useMovementKeys } from "@/game/useMovementKeys";
import { defaultServerUrl } from "@/net/client";
import { useRoom } from "@/net/useRoom";

import { GameCanvas } from "./GameCanvas";
import { NicknameForm } from "./NicknameForm";
import styles from "./RaccoonWorld.module.css";

const wsUrl = defaultServerUrl();

export function RaccoonWorld() {
  const { state, status, join, sendInput } = useRoom(wsUrl);
  useMovementKeys(state.phase === "playing", sendInput);

  if (state.phase !== "playing" || !state.map) {
    return (
      <NicknameForm
        initialNickname={state.nickname}
        joining={state.phase === "joining"}
        status={status}
        rejection={state.rejection}
        onJoin={join}
      />
    );
  }

  const players = Object.values(state.players);
  return (
    <main className={styles.stage}>
      <GameCanvas wsUrl={wsUrl} mapName={state.map} players={state.players} you={state.you} />
      <header className={styles.hud}>
        <strong>Praça</strong>
        <span>
          {state.nickname} · {players.length}/10 guaxinins
        </span>
        {status === "reconnecting" && <span role="status">Reconectando…</span>}
      </header>
      {/* Screen readers and UI tests read who is in the room from here. */}
      <ul className={styles.srOnly} aria-label="Jogadores na Praça">
        {players.map((p) => (
          <li key={p.id}>{p.nickname}</li>
        ))}
      </ul>
    </main>
  );
}
