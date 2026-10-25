import { describe, expect, it } from "vitest";

import { isTypingTarget, KeyTracker } from "./keyboard";

const none = { up: false, down: false, left: false, right: false };

describe("KeyTracker", () => {
  it("reports a change when a movement key goes down or up", () => {
    const k = new KeyTracker();
    expect(k.down("KeyD")).toEqual({ ...none, right: true });
    expect(k.down("KeyW")).toEqual({ ...none, right: true, up: true });
    expect(k.up("KeyD")).toEqual({ ...none, up: true });
    expect(k.up("KeyW")).toEqual(none);
  });

  it("maps W A S D", () => {
    const k = new KeyTracker();
    expect(k.down("KeyW")?.up).toBe(true);
    expect(k.down("KeyA")?.left).toBe(true);
    expect(k.down("KeyS")?.down).toBe(true);
    expect(k.down("KeyD")?.right).toBe(true);
  });

  it("returns null when nothing changed, so input is sent only on change", () => {
    const k = new KeyTracker();
    k.down("KeyD");
    expect(k.down("KeyD")).toBeNull(); // auto-repeat
    expect(k.up("KeyW")).toBeNull(); // was not held
    expect(k.down("KeyJ")).toBeNull(); // not a movement key
    expect(k.down("ArrowRight")).toBeNull(); // the spec says WASD only
  });

  it("releases everything on reset (window blur) and reports it once", () => {
    const k = new KeyTracker();
    k.down("KeyA");
    k.down("KeyS");
    expect(k.reset()).toEqual(none);
    expect(k.reset()).toBeNull();
  });

  it("uses physical key codes, so layouts like AZERTY still move with the WASD positions", () => {
    const k = new KeyTracker();
    expect(k.down("KeyW")).not.toBeNull(); // the key labelled Z on AZERTY
  });
});

describe("isTypingTarget", () => {
  const el = (tagName: string, isContentEditable = false) =>
    ({ tagName, isContentEditable }) as unknown as EventTarget;

  it.each(["INPUT", "TEXTAREA", "SELECT"])("treats %s as typing", (tag) => {
    expect(isTypingTarget(el(tag))).toBe(true);
  });

  it("treats contenteditable as typing", () => {
    expect(isTypingTarget(el("DIV", true))).toBe(true);
  });

  it("treats the page and the canvas as not typing", () => {
    expect(isTypingTarget(el("BODY"))).toBe(false);
    expect(isTypingTarget(el("CANVAS"))).toBe(false);
    expect(isTypingTarget(null)).toBe(false);
  });
});
