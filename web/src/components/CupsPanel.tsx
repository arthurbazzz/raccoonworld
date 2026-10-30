"use client";

import type { CupsState } from "@/game/cups/cupsState";
import { cupsAt, shuffleEndMs, type CupPose } from "@/game/cups/timeline";
import { useFrameTime } from "@/game/cups/useFrameTime";
import type { PlayerId, PlayerState } from "@/net/protocol";

import styles from "./CupsPanel.module.css";

interface Props {
  cups: CupsState;
  you: PlayerId;
  players: Record<PlayerId, PlayerState>;
  onPick: (slot: number) => void;
  onLeave: () => void;
  onPlayAgain: () => void;
}

/** Where the panel is, for the copy, the buttons and the UI tests. */
type Stage = "queued" | "lobby" | "show" | "shuffle" | "choose" | "reveal" | "results";

const REST: CupPose[] = [0, 1, 2].map((slot) => ({ slot, arc: 0, lift: 0 }));
const UP: CupPose[] = [0, 1, 2].map((slot) => ({ slot, arc: 0, lift: 1 }));

/**
 * The three cups panel over the Praça (RW-005, RC19), for whoever is in the
 * lobby, the match or the queue. The Praça stays visible around it.
 */
export function CupsPanel({ cups, you, players, onPick, onLeave, onPlayAgain }: Props) {
  const queued = cups.queued.includes(you);
  const playing = cups.players.includes(you);
  const now = useFrameTime(cups.phase === "lobby" || cups.phase === "playing");
  const name = (id: PlayerId) => players[id]?.nickname ?? "?";

  const round = cups.round;
  const revealed = cups.reveal && round && cups.reveal.round === round.round ? cups.reveal : null;
  const frame = round ? cupsAt(round, now - round.startedAt) : null;

  let stage: Stage;
  if (cups.phase === "results" && playing) stage = "results";
  else if (queued && !(cups.phase === "lobby")) stage = "queued";
  else if (cups.phase === "lobby") stage = "lobby";
  else if (revealed) stage = "reveal";
  else stage = frame?.stage ?? "show";

  const poses = stage === "reveal" ? UP : frame && playing ? frame.cups : REST;
  const nutSlot =
    stage === "reveal" && revealed
      ? revealed.nut
      : frame && frame.cups[frame.nutCup].lift > 0
        ? frame.cups[frame.nutCup].slot
        : null;
  const canPick = playing && stage === "choose" && cups.myPick === null;
  const pickDeadline = round ? round.startedAt + shuffleEndMs(round) + round.pickMs : now;
  const secondsLeft = (until: number) => Math.max(0, Math.ceil((until - now) / 1000));
  const mine = revealed?.picks.find((p) => p.id === you);
  const hitters = revealed?.picks.filter((p) => p.hit).map((p) => name(p.id)) ?? [];

  return (
    <section
      className={
        stage === "lobby" || stage === "queued" ? `${styles.panel} ${styles.waiting}` : styles.panel
      }
      aria-label="Três copos"
      data-stage={stage}
      data-round={round?.round ?? 0}
    >
      <header className={styles.header}>
        <h2 className={styles.title}>Três copos</h2>
        {round && stage !== "lobby" && stage !== "queued" && stage !== "results" && (
          <span className={styles.round}>
            Rodada {round.round} de {round.total}
          </span>
        )}
      </header>

      {stage === "queued" && (
        <p className={styles.message} role="status">
          Partida em andamento, você entra na próxima
        </p>
      )}

      {stage === "lobby" && (
        <>
          <p className={styles.message} role="status">
            A partida começa em {secondsLeft(cups.lobbyEndsAt ?? now)} s
          </p>
          <ul className={styles.names} aria-label="Inscritos">
            {cups.players.map((id) => (
              <li key={id} className={id === you ? styles.you : undefined}>
                {name(id)}
              </li>
            ))}
          </ul>
        </>
      )}

      {(stage === "show" || stage === "shuffle" || stage === "choose" || stage === "reveal") && (
        <>
          <div className={styles.table}>
            {nutSlot !== null && (
              <span
                className={styles.nut}
                style={{ left: `calc(${(nutSlot + 0.5) * (100 / 3)}% - 11px)` }}
                aria-hidden
              />
            )}
            {poses.map((pose, cup) => {
              const slot = Math.round(pose.slot);
              const chosen = cups.myPick !== null && stage !== "shuffle" && slot === cups.myPick;
              return (
                <button
                  key={cup}
                  type="button"
                  className={chosen ? `${styles.cup} ${styles.chosen}` : styles.cup}
                  style={{
                    left: `calc(${(pose.slot + 0.5) * (100 / 3)}% - 28px)`,
                    transform: `translateY(${pose.arc * 8 - pose.lift * 40}px) scale(${1 + pose.arc * 0.08})`,
                    zIndex: 2 + Math.round(pose.arc * 2),
                  }}
                  data-slot={slot}
                  aria-label={`Copo ${slot + 1}`}
                  disabled={!canPick}
                  onClick={() => onPick(slot)}
                >
                  <span className={styles.cupNumber}>{slot + 1}</span>
                </button>
              );
            })}
          </div>

          <p className={styles.message} role="status">
            {stage === "show" && "Olho na noz!"}
            {stage === "shuffle" && "Embaralhando…"}
            {stage === "choose" &&
              (!playing
                ? "Os jogadores estão escolhendo…"
                : cups.myPick === null
                  ? `Escolha um copo: clique ou aperte 1, 2 ou 3 · ${secondsLeft(pickDeadline)} s`
                  : `Você escolheu o copo ${cups.myPick + 1}. Esperando os outros… (${cups.picked.length}/${cups.players.length})`)}
            {stage === "reveal" && (
              <>
                {mine && (
                  <strong className={mine.hit ? styles.hit : styles.miss}>
                    {mine.hit ? "Acertou!" : "Errou!"}
                  </strong>
                )}{" "}
                {hitters.length ? `Acertaram: ${hitters.join(", ")}` : "Ninguém acertou."}
              </>
            )}
          </p>
        </>
      )}

      {stage === "results" && cups.scores && (
        <>
          <p className={styles.message}>Fim de jogo!</p>
          <ol className={styles.scores} aria-label="Placar">
            {cups.scores.map((s) => (
              <li key={s.id} className={s.id === you ? styles.you : undefined}>
                <span>{s.place}º</span> <span>{name(s.id)}</span>{" "}
                <span>
                  {s.hits} de {round?.total ?? 5}
                </span>
              </li>
            ))}
          </ol>
        </>
      )}

      <footer className={styles.actions}>
        {stage === "results" && !queued && (
          <button type="button" className={styles.primary} onClick={onPlayAgain}>
            Jogar de novo
          </button>
        )}
        {stage === "results" && queued && (
          <span className={styles.note}>Você está na próxima partida</span>
        )}
        <button type="button" className={styles.secondary} onClick={onLeave}>
          Sair <kbd>Esc</kbd>
        </button>
      </footer>
    </section>
  );
}
