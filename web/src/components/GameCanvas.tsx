"use client";

import { useEffect, useRef, useState } from "react";

import type { Interpolator } from "@/game/interpolation";
import type { RoomScene } from "@/game/scene";
import { loadTileMap } from "@/game/tilemap";
import type { PlayerId, PlayerState } from "@/net/protocol";

interface Props {
  wsUrl: string;
  mapName: string;
  players: Record<PlayerId, PlayerState>;
  you: PlayerId | null;
  interpolator: Interpolator;
}

/** Hosts the PixiJS scene. PixiJS is loaded client-side only. */
export function GameCanvas({ wsUrl, mapName, players, you, interpolator }: Props) {
  const hostRef = useRef<HTMLDivElement>(null);
  const sceneRef = useRef<RoomScene | null>(null);
  const [error, setError] = useState<string | null>(null);
  const latest = useRef({ players, you });
  latest.current = { players, you };

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
        const created = await RoomScene.create(host, map, interpolator);
        if (cancelled) {
          created.destroy();
          return;
        }
        scene = created;
        sceneRef.current = created;
        created.sync(latest.current.players, latest.current.you);
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e));
      }
    })();

    return () => {
      cancelled = true;
      sceneRef.current = null;
      scene?.destroy();
    };
  }, [wsUrl, mapName, interpolator]);

  useEffect(() => {
    sceneRef.current?.sync(players, you);
  }, [players, you]);

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
