"use client";

import { useCallback, useEffect, useReducer, useRef, useState } from "react";

import { initialRoomState, roomReducer, startJoin, type RoomState } from "@/game/roomState";

import { defaultServerUrl, RoomClient, type Status } from "./client";
import type { Keys, ServerMessage } from "./protocol";

type Action = { kind: "server"; msg: ServerMessage } | { kind: "join"; nickname: string };

function reducer(state: RoomState, action: Action): RoomState {
  return action.kind === "join"
    ? startJoin(state, action.nickname)
    : roomReducer(state, action.msg);
}

export interface Room {
  state: RoomState;
  status: Status | "idle";
  join: (nickname: string) => void;
  sendInput: (keys: Keys) => void;
  sendChat: (text: string) => void;
}

/** Owns one RoomClient for the lifetime of the component. */
export function useRoom(url: string = defaultServerUrl()): Room {
  const [state, dispatch] = useReducer(reducer, initialRoomState);
  const [status, setStatus] = useState<Status | "idle">("idle");
  const clientRef = useRef<RoomClient | null>(null);

  useEffect(() => {
    const client = new RoomClient(url, {
      onMessage: (msg) => dispatch({ kind: "server", msg }),
      onStatus: setStatus,
    });
    clientRef.current = client;
    return () => {
      client.close();
      clientRef.current = null;
    };
  }, [url]);

  const join = useCallback((nickname: string) => {
    dispatch({ kind: "join", nickname });
    clientRef.current?.join(nickname);
  }, []);
  const sendInput = useCallback((keys: Keys) => clientRef.current?.sendInput(keys), []);
  const sendChat = useCallback((text: string) => clientRef.current?.sendChat(text), []);

  return { state, status, join, sendInput, sendChat };
}
