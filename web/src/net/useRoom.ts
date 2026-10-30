"use client";

import { useCallback, useEffect, useReducer, useRef, useState } from "react";

import { ChatBalloons } from "@/game/chat";
import { Interpolator } from "@/game/interpolation";
import { pickLocally } from "@/game/cups/cupsState";
import { initialRoomState, roomReducer, startJoin, type RoomState } from "@/game/roomState";

import { defaultServerUrl, RoomClient, type Status } from "./client";
import type { Keys, ServerMessage } from "./protocol";

type Action =
  | { kind: "server"; msg: ServerMessage; at: number }
  | { kind: "join"; nickname: string }
  | { kind: "pick"; slot: number };

function reducer(state: RoomState, action: Action): RoomState {
  switch (action.kind) {
    case "join":
      return startJoin(state, action.nickname);
    case "pick":
      return { ...state, cups: pickLocally(state.cups, action.slot) };
    default:
      return roomReducer(state, action.msg, action.at);
  }
}

export interface Room {
  state: RoomState;
  status: Status | "idle";
  join: (nickname: string) => void;
  sendInput: (keys: Keys) => void;
  sendChat: (text: string) => void;
  /** Three cups (RW-005): E near the booth, pick a slot, Esc / Sair. */
  joinCups: () => void;
  pickCup: (slot: number) => void;
  leaveCups: () => void;
  /** Smoothed positions for rendering, fed as messages arrive. */
  interpolator: Interpolator;
  /** Chat balloons above raccoons, fed as messages arrive. */
  balloons: ChatBalloons;
}

/** Owns one RoomClient for the lifetime of the component. */
export function useRoom(url: string = defaultServerUrl()): Room {
  const [state, dispatch] = useReducer(reducer, initialRoomState);
  const [status, setStatus] = useState<Status | "idle">("idle");
  const clientRef = useRef<RoomClient | null>(null);
  const [interpolator] = useState(() => new Interpolator());
  const [balloons] = useState(() => new ChatBalloons());

  useEffect(() => {
    const client = new RoomClient(url, {
      onMessage: (msg) => {
        // Timestamp on arrival, outside React, so render batching adds no jitter.
        const now = performance.now();
        interpolator.ingest(msg, now);
        balloons.ingest(msg, now);
        dispatch({ kind: "server", msg, at: now });
      },
      onStatus: setStatus,
    });
    clientRef.current = client;
    return () => {
      client.close();
      clientRef.current = null;
    };
  }, [url, interpolator, balloons]);

  const join = useCallback((nickname: string) => {
    dispatch({ kind: "join", nickname });
    clientRef.current?.join(nickname);
  }, []);
  const sendInput = useCallback((keys: Keys) => clientRef.current?.sendInput(keys), []);
  const sendChat = useCallback((text: string) => clientRef.current?.sendChat(text), []);
  const joinCups = useCallback(() => clientRef.current?.joinCups(), []);
  const leaveCups = useCallback(() => clientRef.current?.leaveCups(), []);
  const round = state.cups.round?.round ?? 0;
  const pickCup = useCallback(
    (slot: number) => {
      dispatch({ kind: "pick", slot });
      clientRef.current?.pickCup(round, slot);
    },
    [round],
  );

  return {
    state,
    status,
    join,
    sendInput,
    sendChat,
    joinCups,
    pickCup,
    leaveCups,
    interpolator,
    balloons,
  };
}
