"use client";

import { roomOf, secondsUntil, type RummyState } from "@/game/rummy/rummyState";
import { useFrameTime } from "@/game/cups/useFrameTime";
import type { PlayerId, PlayerState, RummyMode } from "@/net/protocol";

import styles from "./RummyPanel.module.css";

interface Props {
  rummy: RummyState;
  you: PlayerId;
  players: Record<PlayerId, PlayerState>;
  onCreate: (mode: RummyMode) => void;
  onJoin: (room: string) => void;
  onLeave: () => void;
  onClose: () => void;
}

const MODE_NAMES: Record<RummyMode, string> = { classic: "Clássico", free: "Livre" };
const MODE_HINTS: Record<RummyMode, string> = {
  classic: "a primeira jogada precisa somar 30 pontos",
  free: "pode baixar qualquer combinação desde a primeira vez",
};
const MAX_PLAYERS = 4;

const REJECTIONS: Record<string, string> = {
  too_far: "Chegue mais perto da mesa.",
  full: "Essa sala está cheia.",
  playing: "Essa sala já está jogando.",
  busy: "Você já está em outro jogo.",
  no_room: "Essa sala não existe mais.",
};

/**
 * The room list of the pieces table (RW-006, RN5-RN7), and the lobby of
 * the room we are waiting in.
 */
export function TableRoomsPanel({
  rummy,
  you,
  players,
  onCreate,
  onJoin,
  onLeave,
  onClose,
}: Props) {
  const mine = roomOf(rummy, you);
  const now = useFrameTime(mine?.lobbyEndsAt != null);
  const name = (id: PlayerId) => players[id]?.nickname ?? "?";
  const rejection = rummy.rejection ? (REJECTIONS[rummy.rejection] ?? rummy.rejection) : null;

  if (mine) {
    const left = mine.lobbyEndsAt === null ? null : secondsUntil(mine.lobbyEndsAt, now);
    return (
      <section
        className={`${styles.card} ${styles.rooms}`}
        aria-label="Sala do Rummy das Nozes"
        data-stage="lobby"
      >
        <header className={styles.header}>
          <h2 className={styles.title}>Rummy das Nozes · {MODE_NAMES[mine.mode]}</h2>
          <span className={styles.muted}>
            {mine.players.length}/{MAX_PLAYERS}
          </span>
        </header>
        <ul className={styles.names}>
          {mine.players.map((id) => (
            <li key={id}>{name(id)}</li>
          ))}
        </ul>
        <p className={styles.message} role="status">
          {left === null
            ? "Esperando mais um jogador para começar…"
            : `A partida começa em ${left} s`}
        </p>
        <div className={styles.actions}>
          <button type="button" className={styles.secondary} onClick={onLeave}>
            Sair
          </button>
        </div>
      </section>
    );
  }

  return (
    <section
      className={`${styles.card} ${styles.rooms}`}
      aria-label="Salas do Rummy das Nozes"
      data-stage="list"
    >
      <header className={styles.header}>
        <h2 className={styles.title}>Rummy das Nozes</h2>
        <button type="button" className={styles.close} onClick={onClose} aria-label="Fechar">
          ×
        </button>
      </header>

      {rummy.rooms.length === 0 ? (
        <p className={styles.muted}>Nenhuma sala ainda. Crie uma!</p>
      ) : (
        <ul className={styles.roomList}>
          {rummy.rooms.map((r) => {
            const seats = MAX_PLAYERS - r.players.length;
            const open = r.phase === "lobby" && seats > 0;
            return (
              <li key={r.id} className={styles.roomItem} data-room={r.id} data-phase={r.phase}>
                <div>
                  <strong>{MODE_NAMES[r.mode]}</strong>
                  <span className={styles.muted}> · {r.players.map(name).join(", ")}</span>
                  <div className={styles.muted}>
                    {r.phase === "lobby" ? `${seats} ${seats === 1 ? "vaga" : "vagas"}` : "jogando"}
                  </div>
                </div>
                <button type="button" disabled={!open} onClick={() => onJoin(r.id)}>
                  Entrar
                </button>
              </li>
            );
          })}
        </ul>
      )}

      <div className={styles.create}>
        {(["classic", "free"] as const).map((mode) => (
          <button key={mode} type="button" onClick={() => onCreate(mode)} title={MODE_HINTS[mode]}>
            Criar sala {MODE_NAMES[mode]}
          </button>
        ))}
      </div>
      <p className={styles.hint}>
        <strong>Clássico:</strong> {MODE_HINTS.classic}. <strong>Livre:</strong> {MODE_HINTS.free}.
      </p>
      {rejection && (
        <p className={styles.error} role="alert">
          {rejection}
        </p>
      )}
    </section>
  );
}
