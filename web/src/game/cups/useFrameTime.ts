"use client";

import { useEffect, useState } from "react";

/** Re-renders every animation frame while `on`; returns performance.now(). */
export function useFrameTime(on: boolean): number {
  const [now, setNow] = useState(() => performance.now());
  useEffect(() => {
    if (!on) return;
    let id = requestAnimationFrame(function tick(t) {
      setNow(t);
      id = requestAnimationFrame(tick);
    });
    return () => cancelAnimationFrame(id);
  }, [on]);
  return now;
}
