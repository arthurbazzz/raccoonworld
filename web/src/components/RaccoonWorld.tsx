"use client";

import { useCallback, useState } from "react";

import { nearBooth } from "@/game/cups/booth";
import { cupsRole } from "@/game/cups/cupsState";
import { cupsAt } from "@/game/cups/timeline";
import { useCupsKeys } from "@/game/cups/useCupsKeys";
import { useMotionAttributes } from "@/game/useMotionAttributes";
import { useMovementKeys } from "@/game/useMovementKeys";
import type { Rect, TileMap } from "@/game/tilemap";
import { defaultServerUrl } from "@/net/client";
import { useRoom } from "@/net/useRoom";

import { BoothStatus } from "./BoothStatus";
import { ChatBox } from "./ChatBox";
import { CupsPanel } from "./CupsPanel";
import cupsStyles from "./CupsPanel.module.css";
import { GameCanvas } from "./GameCanvas";
import { NicknameForm } from "./NicknameForm";
import styles from "./RaccoonWorld.module.css";

const wsUrl = defaultServerUrl();

export function RaccoonWorld() {
  const room = useRoom(wsUrl);
  const { state, status, join, sendInput, sendChat, interpolator, balloons } = room;
  const { joinCups, pickCup, leaveCups } = room;
  useMovementKeys(state.phase === "playing", sendInput);
  const motions = useMotionAttributes(interpolator, Object.keys(state.players));

  // Three cups (RW-005): near the booth, E plays; at the booth, Esc leaves.
  const [booth, setBooth] = useState<Rect | null>(null);
  const onMap = useCallback((map: TileMap) => setBooth(map.booth), []);
  const me = state.you ? state.players[state.you] : undefined;
  const role = cupsRole(state.cups, state.you);
  const near = me !== undefined && nearBooth(booth, me.x, me.y);
  const canPick = useCallback(() => {
    const { round, reveal, myPick, players } = state.cups;
    if (!round || reveal?.round === round.round || myPick !== null) return false;
    if (!state.you || !players.includes(state.you)) return false;
    return cupsAt(round, performance.now() - round.startedAt).stage === "choose";
  }, [state.cups, state.you]);
  useCupsKeys({
    canJoin: state.phase === "playing" && near && role === null,
    canLeave: role !== null,
    canPick,
    onJoin: joinCups,
    onLeave: leaveCups,
    onPick: pickCup,
  });

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
        cups={state.cups}
        onMap={onMap}
      />
      <header className={styles.hud}>
        <strong>Praça</strong>
        <span>
          {state.nickname} · {players.length}/10 guaxinins
        </span>
        {status === "reconnecting" && <span role="status">Reconectando…</span>}
      </header>
      <ChatBox onSend={sendChat} />

      {near && role === null && (
        <p className={cupsStyles.prompt} role="status">
          Aperte <kbd>E</kbd> para jogar
        </p>
      )}
      {role !== null && state.you && (
        <CupsPanel
          cups={state.cups}
          you={state.you}
          players={state.players}
          onPick={pickCup}
          onLeave={leaveCups}
          onPlayAgain={joinCups}
        />
      )}

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
            data-cups={cupsRole(state.cups, p.id) ?? undefined}
          >
            {p.nickname}
          </li>
        ))}
      </ul>
      <BoothStatus cups={state.cups} className={styles.srOnly} />
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
