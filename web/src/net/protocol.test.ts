import { describe, expect, it } from "vitest";

import { decodeServerMessage, encodeClientMessage } from "./protocol";

describe("encodeClientMessage", () => {
  it("wraps data in a v1 envelope", () => {
    expect(
      JSON.parse(encodeClientMessage({ type: "room.join", data: { nickname: "rocky" } })),
    ).toEqual({ type: "room.join", v: 1, data: { nickname: "rocky" } });
  });

  it("encodes input with every key", () => {
    const raw = encodeClientMessage({
      type: "player.input",
      data: { up: true, down: false, left: false, right: true, seq: 3 },
    });
    expect(JSON.parse(raw).data).toEqual({
      up: true,
      down: false,
      left: false,
      right: true,
      seq: 3,
    });
  });
});

// These payloads match the Go golden tests in server/internal/adapter/ws/protocol_test.go.
describe("decodeServerMessage", () => {
  it.each([
    [
      '{"type":"room.joined","v":1,"data":{"you":"c1","map":"praca","players":[{"id":"c1","nickname":"rocky","x":480,"y":400}]}}',
      {
        type: "room.joined",
        data: {
          you: "c1",
          map: "praca",
          players: [{ id: "c1", nickname: "rocky", x: 480, y: 400 }],
        },
      },
    ],
    [
      '{"type":"room.join_rejected","v":1,"data":{"reason":"already_joined"}}',
      { type: "room.join_rejected", data: { reason: "already_joined" } },
    ],
    [
      '{"type":"player.joined","v":1,"data":{"id":"c1","nickname":"rocky","x":480,"y":400}}',
      { type: "player.joined", data: { id: "c1", nickname: "rocky", x: 480, y: 400 } },
    ],
    [
      '{"type":"player.left","v":1,"data":{"id":"c1"}}',
      { type: "player.left", data: { id: "c1" } },
    ],
    [
      '{"type":"room.snapshot","v":1,"data":{"tick":42,"players":[{"id":"c2","x":488.5,"y":400}]}}',
      { type: "room.snapshot", data: { tick: 42, players: [{ id: "c2", x: 488.5, y: 400 }] } },
    ],
    [
      '{"type":"chat.message","v":1,"data":{"from":"c1","text":"olá!"}}',
      { type: "chat.message", data: { from: "c1", text: "olá!" } },
    ],
  ])("decodes %s", (raw, want) => {
    expect(decodeServerMessage(raw)).toEqual(want);
  });

  it.each([
    ["not json", "{"],
    ["no type", '{"v":1,"data":{}}'],
    ["unknown type", '{"type":"player.teleport","v":1,"data":{}}'],
    ["client-only type", '{"type":"chat.send","v":1,"data":{"text":"oi"}}'],
    ["other version", '{"type":"player.left","v":2,"data":{"id":"c1"}}'],
    ["missing data", '{"type":"player.left","v":1}'],
  ])("rejects %s", (_name, raw) => {
    expect(decodeServerMessage(raw)).toBeNull();
  });
});
