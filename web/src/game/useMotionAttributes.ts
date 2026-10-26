"use client";

import { useEffect, useState } from "react";

import type { PlayerId } from "@/net/protocol";

import type { Interpolator } from "./interpolation";
import { MOTION_REFRESH_MS, readMotions, sameMotions, type Motions } from "./motionAttributes";

/**
 * Facing and walking state of each player for the accessible list, polled
 * from the interpolator every MOTION_REFRESH_MS. State only changes (and
 * React only re-renders) when someone turns, starts or stops.
 */
export function useMotionAttributes(interpolator: Interpolator, ids: PlayerId[]): Motions {
  const [motions, setMotions] = useState<Motions>({});
  const key = ids.join(",");

  useEffect(() => {
    const list = key === "" ? [] : key.split(",");
    const refresh = () => {
      const next = readMotions(interpolator, list, performance.now());
      setMotions((prev) => (sameMotions(prev, next) ? prev : next));
    };
    refresh();
    const timer = setInterval(refresh, MOTION_REFRESH_MS);
    return () => clearInterval(timer);
  }, [interpolator, key]);

  return motions;
}
