"use client";

import { useCallback, useEffect, useRef, useState, type PointerEvent } from "react";

import { useFrameTime } from "@/game/cups/useFrameTime";
import { moveTile, type DropTarget } from "@/game/rummy/layout";
import { classify, points, REASONS } from "@/game/rummy/melds";
import { secondsUntil, type RummyState } from "@/game/rummy/rummyState";
import type { RummyManifest } from "@/game/rummy/sheet";
import { byColor, byNumber } from "@/game/rummy/tiles";
import type { RummyActions } from "@/net/useRoom";
import type { PlayerId, PlayerState, RummyTurnData } from "@/net/protocol";

import { RummyTile } from "./RummyTile";
import styles from "./RummyPanel.module.css";

interface Props {
  rummy: RummyState;
  you: PlayerId;
  players: Record<PlayerId, PlayerState>;
  actions: RummyActions;
  manifest: RummyManifest | null;
}

interface Drag {
  id: number;
  x: number;
  y: number;
}

const OPENING_POINTS = 30;

const sameMeld = (a: number[], b: number[]) =>
  a.length === b.length && a.every((t, i) => t === b[i]);

/** Where the pointer is over the panel, as a drop target. */
function targetAt(x: number, y: number): DropTarget | null {
  const el = document.elementFromPoint(x, y);
  const zone = el?.closest<HTMLElement>("[data-drop]");
  if (!zone) return null;
  const kind = zone.dataset.drop;
  if (kind === "hand") return { kind: "hand" };
  if (kind === "new") return { kind: "new" };
  if (kind !== "meld") return null;
  const index = Number(zone.dataset.meld);
  // Over a tile: before it on its left half, after it on its right half.
  const tile = el?.closest<HTMLElement>("[data-tile]");
  let before: number | null = null;
  if (tile && zone.contains(tile)) {
    const r = tile.getBoundingClientRect();
    const target = x < r.left + r.width / 2 ? tile : tile.nextElementSibling;
    const id = (target as HTMLElement | null)?.dataset?.tile;
    before = id === undefined ? null : Number(id);
  }
  return { kind: "meld", index, before };
}

function turnMessage(t: RummyTurnData, you: PlayerId, name: (id: PlayerId) => string): string {
  const mine = t.by === you;
  const who = mine ? "Você" : name(t.by);
  const drew = t.drew
    ? `${who} comprou uma peça.`
    : `Monte vazio: ${mine ? "você" : who} passou a vez.`;
  switch (t.result) {
    case "ok":
      return mine ? "Jogada aceita!" : `${who} baixou peças.`;
    case "invalid":
      return `${mine ? "Jogada recusada" : `A jogada de ${who} foi recusada`}: ${REASONS[t.reason] ?? t.reason} ${drew}`;
    case "timeout":
      return `Tempo esgotado. ${drew}`;
    default:
      return drew;
  }
}

/** The match panel of the Rummy das Nozes (RW-006, RN9-RN19, RN26). */
export function RummyPanel({ rummy, you, players, actions, manifest }: Props) {
  const m = rummy.match;
  const results = rummy.results;
  const [sort, setSort] = useState<"color" | "number">("color");
  const [drag, setDrag] = useState<Drag | null>(null);
  const [again, setAgain] = useState(false);
  const now = useFrameTime(m !== null && results === null);
  const name = useCallback((id: PlayerId) => players[id]?.nickname ?? "(saiu)", [players]);

  const myTurn = m !== null && results === null && m.current === you;

  const drop = useCallback(
    (x: number, y: number, id: number) => {
      if (!m || !myTurn) return;
      const target = targetAt(x, y);
      if (!target) return;
      const next = moveTile({ melds: m.melds, hand: m.hand }, id, target, m.startHand);
      if (next) actions.move(m.turn, next.melds, next.hand);
    },
    [m, myTurn, actions],
  );

  // The latest drop, for listeners added when a drag starts.
  const dropRef = useRef(drop);
  useEffect(() => {
    dropRef.current = drop;
  }, [drop]);

  // Listeners go on the window right at pointerdown: a quick drag can end
  // before React would run an effect.
  const startDrag = (e: PointerEvent<HTMLDivElement>, id: number) => {
    if (!myTurn || e.button !== 0) return;
    e.preventDefault();
    setDrag({ id, x: e.clientX, y: e.clientY });
    const onMove = (ev: globalThis.PointerEvent) =>
      setDrag((d) => (d ? { ...d, x: ev.clientX, y: ev.clientY } : d));
    const onUp = (ev: globalThis.PointerEvent) => {
      window.removeEventListener("pointermove", onMove);
      setDrag(null);
      dropRef.current(ev.clientX, ev.clientY, id);
    };
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp, { once: true });
  };

  if (results) {
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

  if (!m) return null;

  const secondsLeft = secondsUntil(m.turnEndsAt, now);
  const hand = [...m.hand].sort(sort === "color" ? byColor : byNumber);
  const opening = m.mode === "classic" && !m.opened.includes(you);
  const newMelds = m.melds.filter((meld) => !m.startMelds.some((s) => sameMeld(s, meld)));
  const openingPoints = newMelds.reduce((sum, meld) => sum + points(meld), 0);
  const tileProps = (id: number) => ({
    id,
    manifest,
    draggable: myTurn,
    dragging: drag?.id === id,
    onPointerDown: startDrag,
  });

  return (
    <section
      className={`${styles.card} ${styles.match}`}
      aria-label="Partida do Rummy das Nozes"
      data-stage={myTurn ? "my-turn" : "watching"}
      data-turn={m.turn}
    >
      <header className={styles.header}>
        <h2 className={styles.title}>
          Rummy das Nozes · {m.mode === "classic" ? "Clássico" : "Livre"}
        </h2>
        <span className={styles.muted}>Monte: {m.pool}</span>
      </header>

      <ul className={styles.seats} aria-label="Jogadores">
        {Object.entries(m.counts).map(([id, count]) => (
          <li
            key={id}
            className={`${styles.seat} ${id === m.current ? styles.current : ""}`}
            data-player={name(id)}
            data-count={count}
          >
            {name(id)}
            {id === you && " (você)"} · {count} {count === 1 ? "peça" : "peças"}
            {m.absent.includes(id) && <span className={styles.absent}> · ausente</span>}
          </li>
        ))}
      </ul>

      <p className={styles.turnLine} role="status">
        <strong>{myTurn ? "Sua vez!" : `Vez de ${name(m.current)}`}</strong>
        <span className={secondsLeft <= 10 ? styles.timerLow : styles.timer}>{secondsLeft} s</span>
      </p>
      {rummy.lastTurn && <p className={styles.message}>{turnMessage(rummy.lastTurn, you, name)}</p>}
      {myTurn && opening && (
        <p className={styles.notice}>
          Abertura: baixe pelo menos {OPENING_POINTS} pontos só com peças da sua mão (
          {openingPoints}/{OPENING_POINTS}).
        </p>
      )}

      <div className={styles.board} aria-label="Mesa">
        {m.melds.map((meld, i) => {
          const c = classify(meld);
          const invalid = "error" in c;
          return (
            <div
              key={i}
              className={`${styles.meld} ${invalid ? styles.invalid : ""}`}
              data-drop="meld"
              data-meld={i}
              data-valid={!invalid}
              title={invalid ? REASONS[c.error] : undefined}
            >
              {meld.map((id) => (
                <RummyTile key={id} {...tileProps(id)} />
              ))}
            </div>
          );
        })}
        {myTurn && (
          <div className={styles.newMeld} data-drop="new">
            Solte aqui para uma nova combinação
          </div>
        )}
        {m.melds.length === 0 && !myTurn && <p className={styles.muted}>A mesa está vazia.</p>}
      </div>

      <div className={styles.handArea} data-drop="hand" aria-label="Sua mão">
        <div className={styles.handHeader}>
          <span>Sua mão ({m.hand.length})</span>
          <span className={styles.sortButtons}>
            <button type="button" aria-pressed={sort === "color"} onClick={() => setSort("color")}>
              Por cor
            </button>
            <button
              type="button"
              aria-pressed={sort === "number"}
              onClick={() => setSort("number")}
            >
              Por número
            </button>
          </span>
        </div>
        <div className={styles.handTiles}>
          {hand.map((id) => (
            <RummyTile key={id} {...tileProps(id)} />
          ))}
        </div>
      </div>

      <div className={styles.actions}>
        {myTurn && (
          <>
            <button
              type="button"
              className={styles.primary}
              onClick={() => actions.turn("end", m.turn)}
            >
              Terminar vez
            </button>
            <button type="button" onClick={() => actions.turn("draw", m.turn)}>
              Comprar
            </button>
            <button type="button" onClick={() => actions.turn("undo", m.turn)}>
              Desfazer
            </button>
          </>
        )}
        <button type="button" className={styles.secondary} onClick={actions.leave}>
          Sair
        </button>
      </div>

      {drag && (
        <div className={styles.ghost} style={{ left: drag.x, top: drag.y }} aria-hidden>
          <RummyTile id={drag.id} manifest={manifest} />
        </div>
      )}
    </section>
  );
}
