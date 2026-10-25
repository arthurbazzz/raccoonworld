import { describe, expect, it } from "vitest";

import { BALLOON_MS, ChatBalloons, MAX_MESSAGE_LENGTH, prepareMessage } from "./chat";

describe("ChatBalloons", () => {
  it("shows a balloon for 5 seconds", () => {
    const b = new ChatBalloons();
    b.ingest({ type: "chat.message", data: { from: "a", text: "olá!" } }, 1000);

    expect(b.textFor("a", 1000)).toBe("olá!");
    expect(b.textFor("a", 1000 + BALLOON_MS - 1)).toBe("olá!");
    expect(b.textFor("a", 1000 + BALLOON_MS)).toBeNull();
    expect(BALLOON_MS).toBe(5000);
  });

  it("replaces the balloon when the same player speaks again", () => {
    const b = new ChatBalloons();
    b.ingest({ type: "chat.message", data: { from: "a", text: "um" } }, 0);
    b.ingest({ type: "chat.message", data: { from: "a", text: "dois" } }, 3000);

    expect(b.textFor("a", 6000)).toBe("dois"); // timer restarted
  });

  it("keeps one balloon per player", () => {
    const b = new ChatBalloons();
    b.ingest({ type: "chat.message", data: { from: "a", text: "oi" } }, 0);
    b.ingest({ type: "chat.message", data: { from: "b", text: "e aí" } }, 0);
    expect(b.textFor("a", 10)).toBe("oi");
    expect(b.textFor("b", 10)).toBe("e aí");
  });

  it("removes a player's balloon when they leave or the room is re-joined", () => {
    const b = new ChatBalloons();
    b.ingest({ type: "chat.message", data: { from: "a", text: "oi" } }, 0);
    b.ingest({ type: "player.left", data: { id: "a" } }, 10);
    expect(b.textFor("a", 20)).toBeNull();

    b.ingest({ type: "chat.message", data: { from: "b", text: "oi" } }, 30);
    b.ingest({ type: "room.joined", data: { you: "c", map: "praca", players: [] } }, 40);
    expect(b.textFor("b", 50)).toBeNull();
  });
});

describe("prepareMessage", () => {
  it("sends the text as typed", () => {
    expect(prepareMessage("  olá, Praça!  ")).toBe("  olá, Praça!  ");
  });

  it("drops blank messages (RN4)", () => {
    expect(prepareMessage("")).toBeNull();
    expect(prepareMessage("   \t")).toBeNull();
  });

  it("drops messages over 120 characters, counting characters, not UTF-16 units", () => {
    expect(MAX_MESSAGE_LENGTH).toBe(120);
    expect(prepareMessage("é".repeat(120))).not.toBeNull();
    expect(prepareMessage("🦝".repeat(120))).not.toBeNull(); // 240 UTF-16 units
    expect(prepareMessage("a".repeat(121))).toBeNull();
  });
});
