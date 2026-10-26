"use client";

import { useMotionAttributes } from "@/game/useMotionAttributes";
import { useMovementKeys } from "@/game/useMovementKeys";
import { defaultServerUrl } from "@/net/client";
import { useRoom } from "@/net/useRoom";

import { ChatBox } from "./ChatBox";
import { GameCanvas } from "./GameCanvas";
import { NicknameForm } from "./NicknameForm";
import styles from "./RaccoonWorld.module.css";

const wsUrl = defaultServerUrl();

export function RaccoonWorld() {
  const { state, status, join, sendInput, sendChat, interpolator, balloons } = useRoom(wsUrl);
  useMovementKeys(state.phase === "playing", sendInput);
  const motions = useMotionAttributes(interpolator, Object.keys(state.players));

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
      <GameCanvas
        wsUrl={wsUrl}
        mapName={state.map}
        players={state.players}
        you={state.you}
        interpolator={interpolator}
        balloons={balloons}
      />
      <header className={styles.hud}>
        <strong>Praça</strong>
        <span>
          {state.nickname} · {players.length}/10 guaxinins
        </span>
        {status === "reconnecting" && <span role="status">Reconectando…</span>}
      </header>
      <ChatBox onSend={sendChat} />

      {/* Screen readers and UI tests read the room from here. */}
      <ul className={styles.srOnly} aria-label="Jogadores na Praça">
        {players.map((p) => (
          // Position: the server's. Facing/moving: what the scene draws (RW-002).
          <li
            key={p.id}
            data-player={p.nickname}
            data-x={p.x}
            data-y={p.y}
            data-facing={motions[p.id]?.facing}
            data-moving={motions[p.id]?.moving}
            data-ring={p.id === state.you}
          >
            {p.nickname}
          </li>
        ))}
      </ul>
      <ol className={styles.srOnly} aria-label="Mensagens" aria-live="polite">
        {state.messages.map((m, i) => (
          <li key={i}>
            {m.nickname}: {m.text}
          </li>
        ))}
      </ol>
    </main>
  );
}
