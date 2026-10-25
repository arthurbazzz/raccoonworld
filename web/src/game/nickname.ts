import type { RejectReason } from "@/net/protocol";

// RN1, mirrored from the server for instant feedback. The server still
// validates every join.
const MIN = 3;
const MAX = 16;
const ALLOWED = /^[A-Za-z0-9_]+$/;

/** Returns why a nickname is invalid, or null when it is fine. */
export function nicknameError(nickname: string): string | null {
  if (nickname === "") return "Digite um apelido.";
  if (nickname.length < MIN || nickname.length > MAX) {
    return `O apelido precisa ter de ${MIN} a ${MAX} caracteres.`;
  }
  if (!ALLOWED.test(nickname)) return "Use só letras, números e _.";
  return null;
}

const rejectMessages: Record<RejectReason, string> = {
  nickname_taken: "Esse apelido já está em uso na Praça. Escolha outro.",
  nickname_invalid: "Apelido inválido. Use de 3 a 16 letras, números ou _.",
  room_full: "A Praça está cheia (10 guaxinins). Tente de novo daqui a pouco.",
  already_joined: "Você já está na Praça.",
};

/** Player-facing text for a room.join_rejected reason. */
export function rejectMessage(reason: RejectReason): string {
  return rejectMessages[reason];
}
