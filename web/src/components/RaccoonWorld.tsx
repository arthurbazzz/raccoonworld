"use client";

import { useRoom } from "@/net/useRoom";

import { NicknameForm } from "./NicknameForm";

export function RaccoonWorld() {
  const { state, status, join } = useRoom();

  if (state.phase !== "playing") {
    return (
      <NicknameForm
        initialNickname={state.nickname}
        joining={state.phase === "joining"}
        status={status}
        rejection={state.rejection}
        onJoin={join}
      />
    );
  }

  // Placeholder until the PixiJS scene lands in T20.
  const players = Object.values(state.players);
  return (
    <main style={{ padding: 16 }}>
      <h1>Praça</h1>
      <p>
        Você entrou como <strong>{state.nickname}</strong>.
      </p>
      <ul aria-label="Jogadores na Praça">
        {players.map((p) => (
          <li key={p.id}>{p.nickname}</li>
        ))}
      </ul>
    </main>
  );
}
