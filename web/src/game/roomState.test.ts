import { describe, expect, it } from "vitest";

import type { ServerMessage } from "@/net/protocol";

import { initialRoomState, roomReducer, type RoomState } from "./roomState";

const rocky = { id: "c1", nickname: "rocky", x: 480, y: 400 };
const bandit = { id: "c2", nickname: "bandit", x: 480, y: 400 };

function apply(state: RoomState, ...msgs: ServerMessage[]): RoomState {
  return msgs.reduce(roomReducer, state);
}

const joining: RoomState = { ...initialRoomState, phase: "joining", nickname: "rocky" };

describe("roomReducer", () => {
  it("enters the room on room.joined", () => {
    const s = apply(joining, {
      type: "room.joined",
      data: { you: "c1", map: "praca", players: [bandit, rocky] },
    });
    expect(s.phase).toBe("playing");
    expect(s.you).toBe("c1");
    expect(s.map).toBe("praca");
    expect(s.players).toEqual({ c1: rocky, c2: bandit });
    expect(s.rejection).toBeNull();
  });

  it("goes back to the nickname screen on rejection", () => {
    const s = apply(joining, { type: "room.join_rejected", data: { reason: "room_full" } });
    expect(s.phase).toBe("nickname");
    expect(s.rejection).toBe("room_full");
  });

  it("ignores already_joined: the player is still in the room", () => {
    const playing = apply(joining, {
      type: "room.joined",
      data: { you: "c1", map: "praca", players: [rocky] },
    });
    const s = apply(playing, { type: "room.join_rejected", data: { reason: "already_joined" } });
    expect(s).toBe(playing);
  });

  it("tracks arrivals and departures", () => {
    let s = apply(joining, {
      type: "room.joined",
      data: { you: "c1", map: "praca", players: [rocky] },
    });
    s = apply(s, { type: "player.joined", data: rocky }); // own arrival: no change
    s = apply(s, { type: "player.joined", data: bandit });
    expect(Object.keys(s.players).sort()).toEqual(["c1", "c2"]);

    s = apply(s, { type: "player.left", data: { id: "c2" } });
    expect(Object.keys(s.players)).toEqual(["c1"]);
  });

  it("moves players on snapshots and ignores unknown ids", () => {
    let s = apply(joining, {
      type: "room.joined",
      data: { you: "c1", map: "praca", players: [rocky, bandit] },
    });
    s = apply(s, {
      type: "room.snapshot",
      data: {
        tick: 7,
        players: [
          { id: "c2", x: 500, y: 390 },
          { id: "ghost", x: 1, y: 1 },
        ],
      },
    });
    expect(s.players.c2).toEqual({ ...bandit, x: 500, y: 390 });
    expect(s.players.c1).toEqual(rocky);
    expect(s.players.ghost).toBeUndefined();
    expect(s.tick).toBe(7);
  });

  it("keeps the latest chat messages with the sender's nickname", () => {
    let s = apply(joining, {
      type: "room.joined",
      data: { you: "c1", map: "praca", players: [rocky, bandit] },
    });
    s = apply(s, { type: "chat.message", data: { from: "c1", text: "olá!" } });
    for (let i = 0; i < 30; i++) {
      s = apply(s, { type: "chat.message", data: { from: "c2", text: `msg ${i}` } });
    }
    expect(s.messages).toHaveLength(20);
    expect(s.messages.at(-1)).toEqual({ from: "c2", nickname: "bandit", text: "msg 29" });
  });

  it("ignores room events before joining", () => {
    const s = apply(joining, { type: "player.joined", data: bandit });
    expect(s).toBe(joining);
  });
});
