"use client";

import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
} from "react";

import { useFrameTime } from "@/game/cups/useFrameTime";
import {
  cutMeld,
  joinMelds,
  moveMeld,
  moveTiles,
  MELD_ROWS,
  TABLE_COLS,
  TABLE_ROWS,
  TILE_COLS,
  type DropTarget,
  type Layout,
} from "@/game/rummy/layout";
import { checkDraft, describeIssue } from "@/game/rummy/melds";
import { placeOnRack, slotsPerRow, sortRack, syncRack, type Rack } from "@/game/rummy/rack";
import { BALLOON_MS, secondsUntil, type RummyState } from "@/game/rummy/rummyState";
import { boxFrom, dragged, tilesInBox, toggle, type Box } from "@/game/rummy/selection";
import type { RummyManifest } from "@/game/rummy/sheet";
import type { PlayerId, PlayerState, RummyIssue, RummyTurnData } from "@/net/protocol";
import type { RummyActions } from "@/net/useRoom";

import { RummyResults } from "./RummyResults";
import { RummyTile } from "./RummyTile";
import { TableChat } from "./TableChat";
import styles from "./RummyTable.module.css";

interface Props {
  rummy: RummyState;
  you: PlayerId;
  players: Record<PlayerId, PlayerState>;
  actions: RummyActions;
  manifest: RummyManifest | null;
}

type Drag =
  | { kind: "tiles"; ids: number[]; x: number; y: number; dx: number; dy: number }
  | { kind: "meld"; index: number; x: number; y: number; dx: number; dy: number };

/** From this width the chat sits beside the table instead of over it. */
const WIDE = 1100;

/** How far the pointer moves before a press becomes a drag, in pixels. */
const DRAG_START = 4;

const sameMeld = (a: number[], b: number[]) =>
  a.length === b.length && a.every((t, i) => t === b[i]);

function turnMessage(t: RummyTurnData, you: PlayerId, name: (id: PlayerId) => string): string {
  const mine = t.by === you;
  const who = mine ? "Você" : name(t.by);
  switch (t.result) {
    case "ok":
      return mine ? "Jogada aceita!" : `${who} baixou peças.`;
    case "timeout":
      return t.drew
        ? `Tempo esgotado: ${mine ? "você comprou" : `${who} comprou`} uma peça.`
        : "Tempo esgotado.";
    default:
      return t.drew ? `${who} comprou uma peça.` : `Monte vazio: ${who} passou a vez.`;
  }
}

/** Where other players sit around the table, by how many there are (RN28). */
const SEATS: Record<number, ("top" | "left" | "right")[]> = {
  1: ["top"],
  2: ["left", "right"],
  3: ["left", "top", "right"],
};

/**
 * The match of the Rummy das Nozes as a full-screen table (RW-006, Emenda
 * A: RN28-RN39): the felt with the melds, the others around it, our tile
 * rack below and the table chat.
 */
export function RummyTableScene({ rummy, you, players, actions, manifest }: Props) {
  const m = rummy.match;
  const results = rummy.results;
  const now = useFrameTime(m !== null && results === null);
  const name = useCallback((id: PlayerId) => players[id]?.nickname ?? "(saiu)", [players]);
  const myTurn = m !== null && results === null && m.current === you;

  // --- the rack (RN30): local slots, kept in step with our hand ---
  const [rack, setRack] = useState<Rack>([]);
  const hand = m?.hand;
  useEffect(() => {
    if (hand) setRack((r) => syncRack(r, hand));
  }, [hand]);

  // --- selection (RN35) ---
  const [selected, setSelected] = useState<number[]>([]);
  useEffect(() => {
    // Esc clears the marks before it would leave the table.
    const onKey = (e: KeyboardEvent) => {
      if (e.code === "Escape" && selected.length > 0) {
        e.stopImmediatePropagation();
        setSelected([]);
      }
    };
    window.addEventListener("keydown", onKey, { capture: true });
    return () => window.removeEventListener("keydown", onKey, { capture: true });
  }, [selected.length]);
  useEffect(() => {
    if (!myTurn) setSelected((s) => (s.length ? s.filter((t) => hand?.includes(t)) : s));
  }, [myTurn, hand]);

  // --- the felt grid: cell size from the felt's width ---
  const boardRef = useRef<HTMLDivElement>(null);
  const [cell, setCell] = useState(18);
  useLayoutEffect(() => {
    const el = boardRef.current;
    if (!el) return;
    const ro = new ResizeObserver(() => setCell(el.clientWidth / TABLE_COLS));
    ro.observe(el);
    return () => ro.disconnect();
  }, [m !== null]); // eslint-disable-line react-hooks/exhaustive-deps

  const layout: Layout | null = m ? { melds: m.melds, pos: m.pos, hand: m.hand } : null;
  const opening = m !== null && m.mode === "classic" && !m.opened.includes(you);

  const send = useCallback(
    (next: Layout | null) => {
      if (next && m && myTurn) actions.move(m.turn, next);
    },
    [m, myTurn, actions],
  );

  // Live issues while we arrange the table (RN12a), and the server's refusal.
  const live: RummyIssue[] = useMemo(
    () => (m && myTurn ? checkDraft(m.melds, m.startMelds, m.startHand, opening) : []),
    [m, myTurn, opening],
  );
  const refusal = rummy.refusal;
  const issues = useMemo(() => refusal ?? [], [refusal]);
  const blamed = useMemo(() => {
    const out = new Set<number>();
    for (const is of [...live, ...issues]) {
      const tiles = is.tiles.length > 0 ? is.tiles : is.meld >= 0 ? (m?.melds[is.meld] ?? []) : [];
      tiles.forEach((t) => out.add(t));
    }
    return out;
  }, [live, issues, m]);
  const meldIssue = (i: number) =>
    live.find((is) => is.meld === i && is.code !== "opening_used_table") ??
    live.find((is) => is.meld === i);

  // --- drag and drop ---
  const [drag, setDrag] = useState<Drag | null>(null);
  const [box, setBox] = useState<Box | null>(null);

  const order = useMemo(
    () => [...(m?.melds.flat() ?? []), ...rack.filter((t): t is number => t !== null)],
    [m, rack],
  );

  /** The table cell under a viewport point, for a meld grabbed `dx, dy` from its corner. */
  const cellAt = (x: number, y: number, dx: number, dy: number): [number, number] => {
    const r = boardRef.current!.getBoundingClientRect();
    return [Math.round((x - dx - r.left) / cell), Math.round((y - dy - r.top) / cell)];
  };

  const dropTiles = (ids: number[], x: number, y: number, dx: number, dy: number) => {
    if (!m || !layout) return;
    const el = document.elementFromPoint(x, y) as HTMLElement | null;
    const slot = el?.closest<HTMLElement>("[data-slot]");
    if (slot) {
      const at = Number(slot.dataset.slot);
      const inHand = ids.every((t) => m.hand.includes(t));
      if (!inHand) {
        if (!myTurn) return;
        const next = moveTiles(layout, ids, { kind: "hand" }, m.startHand);
        if (!next) return;
        send(next);
      }
      setRack((r) => placeOnRack(syncRack(r, [...m.hand, ...ids]), ids, at));
      return;
    }
    if (!myTurn) return;
    if (opening && ids.some((t) => !m.startHand.includes(t))) return; // RN36
    const meldEl = el?.closest<HTMLElement>("[data-meld]");
    let target: DropTarget | null = null;
    if (meldEl) {
      const index = Number(meldEl.dataset.meld);
      const tile = el?.closest<HTMLElement>("[data-tile]");
      let before: number | null = null;
      if (tile && meldEl.contains(tile)) {
        const r = tile.getBoundingClientRect();
        const t = x < r.left + r.width / 2 ? tile : (tile.nextElementSibling as HTMLElement | null);
        const id = t?.dataset?.tile;
        before = id === undefined ? null : Number(id);
      }
      target = { kind: "meld", index, before };
    } else if (el?.closest("[data-felt]")) {
      target = { kind: "new", at: cellAt(x, y, dx, dy) };
    }
    if (!target) return;
    send(moveTiles(layout, ids, target, m.startHand));
    setSelected([]);
  };

  const dropMeld = (index: number, x: number, y: number, dx: number, dy: number) => {
    if (!layout || !m || !myTurn) return;
    const el = document.elementFromPoint(x, y) as HTMLElement | null;
    const other = el?.closest<HTMLElement>("[data-meld]");
    if (other && Number(other.dataset.meld) !== index) {
      if (opening) return; // RN36: no joining table melds before opening
      const to = Number(other.dataset.meld);
      const r = other.getBoundingClientRect();
      send(joinMelds(layout, index, to, x < r.left + r.width / 2 ? "start" : "end"));
      return;
    }
    if (!el?.closest("[data-felt]")) return;
    send(moveMeld(layout, index, cellAt(x, y, dx, dy)));
  };

  // Listeners go on the window at pointerdown: a quick drag can end before
  // React would run an effect.
  const follow = (
    e: ReactPointerEvent,
    begin: (x: number, y: number) => Drag | null,
    end: (d: Drag, x: number, y: number) => void,
    click?: () => void,
  ) => {
    const sx = e.clientX;
    const sy = e.clientY;
    let current: Drag | null = null;
    const onMove = (ev: PointerEvent) => {
      if (!current) {
        if (Math.hypot(ev.clientX - sx, ev.clientY - sy) < DRAG_START) return;
        current = begin(ev.clientX, ev.clientY);
        if (!current) return;
      }
      current = { ...current, x: ev.clientX, y: ev.clientY };
      setDrag(current);
    };
    const onUp = (ev: PointerEvent) => {
      window.removeEventListener("pointermove", onMove);
      setDrag(null);
      if (current) end(current, ev.clientX, ev.clientY);
      else click?.();
    };
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp, { once: true });
  };

  const onTileDown = (e: ReactPointerEvent<HTMLDivElement>, id: number) => {
    if (e.button !== 0 || !m) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.shiftKey) {
      setSelected((s) => toggle(s, id));
      return;
    }
    const onTable = !m.hand.includes(id);
    if (onTable && !myTurn) return;
    const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const dx = e.clientX - r.left;
    const dy = e.clientY - r.top;
    follow(
      e,
      (x, y) => ({ kind: "tiles", ids: dragged(selected, id, order), x, y, dx, dy }),
      (d, x, y) => d.kind === "tiles" && dropTiles(d.ids, x, y, d.dx, d.dy),
      // A click without dragging marks or unmarks the tile (RN35).
      () => setSelected((s) => toggle(s, id)),
    );
  };

  const onHandleDown = (e: ReactPointerEvent<HTMLDivElement>, index: number) => {
    if (e.button !== 0 || !myTurn || !m) return;
    e.preventDefault();
    e.stopPropagation();
    const r = (e.currentTarget.parentElement as HTMLElement).getBoundingClientRect();
    const dx = e.clientX - r.left;
    const dy = e.clientY - r.top;
    follow(
      e,
      (x, y) => ({ kind: "meld", index, x, y, dx, dy }),
      (d, x, y) => d.kind === "meld" && dropMeld(d.index, x, y, d.dx, d.dy),
    );
  };

  // The selection box on an empty part of the felt or the rack (RN35).
  const rackRef = useRef<HTMLDivElement>(null);
  const onFeltDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!myTurn) return;
    onBoxDown(e, boardRef.current);
  };
  const onBoxDown = (e: ReactPointerEvent<HTMLDivElement>, area: HTMLElement | null) => {
    if (e.button !== 0 || !area) return;
    if ((e.target as HTMLElement).closest("[data-meld], [data-tile]")) return;
    e.preventDefault();
    const sx = e.clientX;
    const sy = e.clientY;
    const onMove = (ev: PointerEvent) => setBox(boxFrom(sx, sy, ev.clientX, ev.clientY));
    const onUp = (ev: PointerEvent) => {
      window.removeEventListener("pointermove", onMove);
      setBox(null);
      const b = boxFrom(sx, sy, ev.clientX, ev.clientY);
      if (b.right - b.left < DRAG_START && b.bottom - b.top < DRAG_START) {
        setSelected([]);
        return;
      }
      const tiles = [...area.querySelectorAll<HTMLElement>("[data-tile]")].map((t) => {
        const r = t.getBoundingClientRect();
        return {
          id: Number(t.dataset.tile),
          rect: { left: r.left, top: r.top, right: r.right, bottom: r.bottom },
        };
      });
      const inBox = tilesInBox(b, tiles).filter((t) => !opening || m!.startHand.includes(t));
      setSelected((s) => (ev.shiftKey ? [...new Set([...s, ...inBox])] : inBox));
    };
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp, { once: true });
  };

  // Enter focuses the table chat, as it does the Praça chat.
  // On small screens the chat starts closed, so it does not cover the table.
  const [chatOpen, setChatOpen] = useState(
    () => typeof window === "undefined" || window.innerWidth >= WIDE,
  );

  if (results) {
    return <RummyResults results={results} you={you} name={name} actions={actions} />;
  }
  if (!m || !layout) return null;

  const msLeft = Math.max(0, m.turnEndsAt - Math.max(now, performance.now()));
  const secondsLeft = secondsUntil(m.turnEndsAt, now);
  const fraction = m.turnTotal > 0 ? Math.min(1, msLeft / m.turnTotal) : 0;
  const placed = m.hand.length < m.startHand.length;
  const others = Object.keys(m.counts).filter((id) => id !== you);
  // Seat the others in turn order, starting after us.
  const seatsAt = SEATS[others.length] ?? SEATS[3];
  const now2 = performance.now();
  const balloonOf = (id: PlayerId) => {
    for (let i = rummy.chat.length - 1; i >= 0; i--) {
      const c = rummy.chat[i];
      if (c.from === id) return now2 - c.at < BALLOON_MS ? c.text : null;
    }
    return null;
  };
  const tileH = cell * MELD_ROWS;
  const per = slotsPerRow(m.hand.length);
  const rackH = Math.min(64, Math.max(44, tileH * 1.15));

  return (
    <section
      className={styles.scene}
      aria-label="Partida do Rummy das Nozes"
      data-stage={myTurn ? "my-turn" : "watching"}
      data-turn={m.turn}
    >
      <header className={styles.bar}>
        <h2 className={styles.title}>
          Rummy das Nozes · {m.mode === "classic" ? "Clássico" : "Livre"}
        </h2>
        <p className={styles.status} role="status">
          <strong>{myTurn ? "Sua vez!" : `Vez de ${name(m.current)}`}</strong>{" "}
          <span className={secondsLeft <= 10 ? styles.timerLow : styles.timer}>
            {secondsLeft} s
          </span>
          {rummy.lastTurn && (
            <span className={styles.lastTurn}> · {turnMessage(rummy.lastTurn, you, name)}</span>
          )}
        </p>
        <button type="button" className={styles.leave} onClick={actions.leave}>
          Sair
        </button>
      </header>

      <div className={styles.table}>
        {others.map((id, i) => (
          <Seat
            key={id}
            where={seatsAt[i]}
            name={name(id)}
            count={m.counts[id]}
            current={m.current === id}
            absent={m.absent.includes(id)}
            fraction={m.current === id ? fraction : 0}
            low={secondsLeft <= 10}
            balloon={balloonOf(id)}
          />
        ))}

        <div className={styles.felt} data-felt onPointerDown={onFeltDown}>
          <div
            ref={boardRef}
            className={styles.board}
            style={{ aspectRatio: `${TABLE_COLS} / ${TABLE_ROWS}` }}
            aria-label="Mesa"
          >
            {m.melds.length === 0 && (
              <p className={styles.empty}>
                {myTurn ? "Arraste peças do suporte para cá" : "A mesa está vazia"}
              </p>
            )}
            {m.melds.map((meld, i) => {
              const [x, y] = m.pos[i] ?? [0, 0];
              const issue = meldIssue(i);
              const isNew = !m.startMelds.some((s) => sameMeld(s, meld));
              return (
                <div
                  key={`${meld[0]}-${i}`}
                  className={[
                    styles.meld,
                    issue ? styles.invalid : "",
                    myTurn && isNew ? styles.fresh : "",
                    drag?.kind === "meld" && drag.index === i ? styles.lifted : "",
                  ].join(" ")}
                  style={{ left: x * cell, top: y * cell, height: tileH }}
                  data-meld={i}
                  data-valid={!issue}
                >
                  <div
                    className={styles.handle}
                    style={{ width: cell }}
                    onPointerDown={(e) => onHandleDown(e, i)}
                    title={myTurn ? "Arraste para mover a combinação" : undefined}
                    aria-label="Mover combinação"
                  />
                  {meld.map((id, k) => (
                    <div
                      key={id}
                      className={styles.slotOnTable}
                      style={{ width: cell * TILE_COLS }}
                    >
                      {k > 0 && myTurn && !opening && (
                        <button
                          type="button"
                          className={styles.cut}
                          aria-label={`Cortar antes da peça ${k + 1}`}
                          title="Cortar aqui"
                          onPointerDown={(e) => e.stopPropagation()}
                          onClick={() => send(cutMeld(layout, i, k))}
                        />
                      )}
                      <RummyTile
                        id={id}
                        manifest={manifest}
                        height={tileH}
                        draggable={myTurn}
                        dragging={drag?.kind === "tiles" && drag.ids.includes(id)}
                        marked={selected.includes(id)}
                        blamed={blamed.has(id)}
                        onPointerDown={myTurn ? onTileDown : undefined}
                      />
                    </div>
                  ))}
                  {issue && <p className={styles.meldError}>{describeIssue(issue)}</p>}
                </div>
              );
            })}
          </div>
          <div className={styles.pool} aria-label={`Monte: ${m.pool} peças`}>
            <span className={styles.poolStack} />
            <span>{m.pool}</span>
          </div>
        </div>
      </div>

      {(issues.length > 0 || (myTurn && opening)) && (
        <div className={styles.problems} role="alert">
          {issues.length > 0 && <strong>Jogada recusada. Corrija e tente de novo:</strong>}
          {issues.length > 0 && (
            <ul>
              {issues.map((is, k) => (
                <li key={k}>{describeIssue(is)}</li>
              ))}
            </ul>
          )}
          {issues.length === 0 && opening && (
            <span>Abertura: baixe pelo menos 30 pontos só com peças do seu suporte.</span>
          )}
        </div>
      )}

      <footer className={styles.bottom}>
        <Seat
          where="me"
          name={`${name(you)} (você)`}
          count={m.hand.length}
          current={myTurn}
          absent={false}
          fraction={myTurn ? fraction : 0}
          low={secondsLeft <= 10}
          balloon={balloonOf(you)}
        />
        <div className={styles.rackWrap}>
          <div
            ref={rackRef}
            className={styles.rack}
            onPointerDown={(e) => onBoxDown(e, rackRef.current)}
            style={{ gridTemplateColumns: `repeat(${per}, ${rackH * 0.66}px)` }}
            aria-label="Seu suporte"
          >
            {rack.map((id, slot) => (
              <div
                key={slot}
                className={styles.rackSlot}
                data-slot={slot}
                style={{ height: rackH }}
              >
                {id !== null && (
                  <RummyTile
                    id={id}
                    manifest={manifest}
                    height={rackH - 4}
                    draggable
                    dragging={drag?.kind === "tiles" && drag.ids.includes(id)}
                    marked={selected.includes(id)}
                    onPointerDown={onTileDown}
                  />
                )}
              </div>
            ))}
          </div>
          <div className={styles.sort}>
            <button type="button" onClick={() => setRack(sortRack(m.hand, "color"))}>
              Por cor
            </button>
            <button type="button" onClick={() => setRack(sortRack(m.hand, "number"))}>
              Por número
            </button>
            {selected.length > 0 && (
              <span className={styles.marks}>
                {selected.length} marcada{selected.length > 1 ? "s" : ""} · Esc desmarca
              </span>
            )}
          </div>
        </div>
        <div className={styles.actions}>
          {myTurn && (
            <>
              <button
                type="button"
                className={styles.primary}
                disabled={!placed}
                title={placed ? undefined : "Baixe alguma peça antes, ou clique em Comprar"}
                onClick={() => actions.turn("end", m.turn)}
              >
                Terminar vez
              </button>
              <button type="button" onClick={() => actions.turn("draw", m.turn)}>
                Comprar
              </button>
              <button
                type="button"
                onClick={() => {
                  actions.turn("undo", m.turn);
                  setSelected([]);
                }}
              >
                Desfazer
              </button>
            </>
          )}
        </div>
      </footer>

      <TableChat
        open={chatOpen}
        onToggle={() => setChatOpen((o) => !o)}
        lines={rummy.chat}
        name={name}
        you={you}
        onSend={actions.say}
      />

      {box && <SelectionBox box={box} />}
      {drag?.kind === "tiles" && (
        <div
          className={styles.ghost}
          style={{ left: drag.x - drag.dx, top: drag.y - drag.dy }}
          aria-hidden
        >
          {drag.ids.map((id) => (
            <RummyTile key={id} id={id} manifest={manifest} height={tileH} />
          ))}
        </div>
      )}
      {drag?.kind === "meld" && m.melds[drag.index] && (
        <div
          className={styles.ghost}
          style={{ left: drag.x - drag.dx, top: drag.y - drag.dy }}
          aria-hidden
        >
          <span style={{ width: cell }} />
          {m.melds[drag.index].map((id) => (
            <RummyTile key={id} id={id} manifest={manifest} height={tileH} />
          ))}
        </div>
      )}
    </section>
  );
}

interface SeatProps {
  where: "top" | "left" | "right" | "me";
  name: string;
  count: number;
  current: boolean;
  absent: boolean;
  /** Part of the turn left, 1 to 0, for the ring (RN29). */
  fraction: number;
  low: boolean;
  balloon: string | null;
}

const RING_R = 36;
const RING_C = 2 * Math.PI * RING_R;

/** A player around the table: avatar, ring of the turn, name, tiles and chat balloon. */
function Seat({ where, name, count, current, absent, fraction, low, balloon }: SeatProps) {
  return (
    <div
      className={`${styles.seat} ${styles[where]} ${current ? styles.current : ""}`}
      data-player={name}
      data-count={count}
    >
      <div className={styles.avatarWrap}>
        <svg className={styles.ring} viewBox="0 0 80 80" aria-hidden>
          <circle cx="40" cy="40" r={RING_R} className={styles.ringBack} />
          {current && (
            <circle
              cx="40"
              cy="40"
              r={RING_R}
              className={low ? styles.ringLow : styles.ringOn}
              strokeDasharray={RING_C}
              strokeDashoffset={RING_C * (1 - fraction)}
              transform="rotate(-90 40 40)"
            />
          )}
        </svg>
        <div className={styles.avatar} />
        {balloon && (
          <p className={styles.balloon} role="status">
            {balloon}
          </p>
        )}
      </div>
      <div className={styles.nameplate}>
        <strong>{name}</strong>
        <span className={styles.count}>
          <span className={styles.back} aria-hidden /> {count}
        </span>
        {absent && <span className={styles.absent}>ausente</span>}
      </div>
    </div>
  );
}

function SelectionBox({ box }: { box: Box }) {
  return (
    <div
      className={styles.selectBox}
      style={{
        left: box.left,
        top: box.top,
        width: box.right - box.left,
        height: box.bottom - box.top,
      }}
    />
  );
}
