import { describe, expect, it } from "vitest";

import { facingFrom, MIN_STEP_PX, type Facing } from "./motion";

// Screen coordinates: x grows to the right, y grows downward.
describe("facingFrom (RV1)", () => {
  it.each<[string, number, number, Facing]>([
    ["right", 8, 0, "right"],
    ["left", -8, 0, "left"],
    ["down", 0, 8, "down"],
    ["up", 0, -8, "up"],
  ])("faces %s for a straight move", (_name, dx, dy, want) => {
    expect(facingFrom(dx, dy, "down")).toBe(want);
  });

  it.each<[string, number, number, Facing]>([
    ["up-right", 5.66, -5.66, "right"],
    ["down-right", 5.66, 5.66, "right"],
    ["up-left", -5.66, -5.66, "left"],
    ["down-left", -5.66, 5.66, "left"],
  ])("uses the side on a diagonal (%s)", (_name, dx, dy, want) => {
    expect(facingFrom(dx, dy, "up")).toBe(want);
  });

  it("uses the side even when the horizontal part is small", () => {
    expect(facingFrom(1, -7, "down")).toBe("right");
  });

  it.each<Facing>(["down", "up", "left", "right"])("keeps facing %s when not moving", (prev) => {
    expect(facingFrom(0, 0, prev)).toBe(prev);
  });

  it("ignores floating point noise below MIN_STEP_PX", () => {
    expect(MIN_STEP_PX).toBe(0.01);
    expect(facingFrom(0.001, -0.004, "left")).toBe("left");
    // Vertical move with horizontal noise: still vertical.
    expect(facingFrom(0.001, 8, "left")).toBe("down");
  });
});
