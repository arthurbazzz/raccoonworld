"use client";

import { useEffect, useRef, useState } from "react";

import type { ChatBalloons } from "@/game/chat";
import type { CupsState } from "@/game/cups/cupsState";
import type { RummyState } from "@/game/rummy/rummyState";
import type { Interpolator } from "@/game/interpolation";
import type { RoomScene } from "@/game/scene";
import { loadTileMap, type TileMap } from "@/game/tilemap";
import type { PlayerId, PlayerState } from "@/net/protocol";

interface Props {
  wsUrl: string;
  mapName: string;
  players: Record<PlayerId, PlayerState>;
  you: PlayerId | null;
  interpolator: Interpolator;
  balloons: ChatBalloons;
  cups: CupsState;
  rummy: RummyState;
  /** Called once the map is loaded (the world needs the booth, RW-005). */
  onMap?: (map: TileMap) => void;
}

/** Hosts the PixiJS scene. PixiJS is loaded client-side only. */
export function GameCanvas({
  wsUrl,
  mapName,
  players,
  you,
  interpolator,
  balloons,
  cups,
  rummy,
  onMap,
}: Props) {
  const hostRef = useRef<HTMLDivElement>(null);
  const sceneRef = useRef<RoomScene | null>(null);
  const [error, setError] = useState<string | null>(null);
  const latest = useRef({ players, you, cups, rummy, onMap });
  latest.current = { players, you, cups, rummy, onMap };

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;
    let cancelled = false;
    let scene: RoomScene | null = null;

    (async () => {
      try {
        const [{ RoomScene }, map] = await Promise.all([
          import("@/game/scene"),
          loadTileMap(wsUrl, mapName),
        ]);
        if (cancelled) return;
        latest.current.onMap?.(map);
        const created = await RoomScene.create(host, map, interpolator, balloons);
        if (cancelled) {
          created.destroy();
          return;
        }
        scene = created;
        sceneRef.current = created;
        created.sync(latest.current.players, latest.current.you);
        created.syncCups(latest.current.cups);
        created.syncRummy(latest.current.rummy);
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e));
      }
    })();

    return () => {
      cancelled = true;
      sceneRef.current = null;
      scene?.destroy();
    };
  }, [wsUrl, mapName, interpolator, balloons]);

  useEffect(() => {
    sceneRef.current?.sync(players, you);
  }, [players, you]);

  useEffect(() => {
    sceneRef.current?.syncCups(cups);
  }, [cups]);

  useEffect(() => {
    sceneRef.current?.syncRummy(rummy);
  }, [rummy]);

  return (
    <div ref={hostRef} style={{ position: "absolute", inset: 0 }}>
      {error && (
        <p role="alert" style={{ padding: 16 }}>
          Não foi possível carregar a Praça: {error}
        </p>
      )}
    </div>
  );
}
