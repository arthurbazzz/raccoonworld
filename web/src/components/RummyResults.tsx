"use client";

import { useState } from "react";

import type { PlayerId, RummyResultsData } from "@/net/protocol";
import type { RummyActions } from "@/net/useRoom";

import styles from "./RummyPanel.module.css";

interface Props {
  results: RummyResultsData;
  you: PlayerId;
  name: (id: PlayerId) => string;
  actions: RummyActions;
}

/** The scoreboard at the end of a match (RN19), with Jogar de novo (RN7). */
export function RummyResults({ results, you, name, actions }: Props) {
  const [again, setAgain] = useState(false);
  return (
    <section
      className={`${styles.card} ${styles.match}`}
      aria-label="Placar do Rummy das Nozes"
      data-stage="results"
    >
      <header className={styles.header}>
        <h2 className={styles.title}>Fim da partida</h2>
      </header>
      <p className={styles.winner} role="status">
        {results.winner === null
          ? "Empate!"
          : results.winner === you
            ? "Você ganhou! 🎉"
            : `${name(results.winner)} ganhou!`}
      </p>
      <ol className={styles.scores}>
        {results.scores.map((s) => (
          <li key={s.id} data-player={name(s.id)} data-points={s.points}>
            <span>
              {s.place}º {name(s.id)}
            </span>
            <span>{s.points} pontos na mão</span>
          </li>
        ))}
      </ol>
      <div className={styles.actions}>
        {again ? (
          <span className={styles.muted}>Esperando a próxima partida…</span>
        ) : (
          <button
            type="button"
            className={styles.primary}
            onClick={() => {
              setAgain(true);
              actions.again();
            }}
          >
            Jogar de novo
          </button>
        )}
        <button type="button" className={styles.secondary} onClick={actions.leave}>
          Sair
        </button>
      </div>
    </section>
  );
}
