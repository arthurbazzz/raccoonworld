import { describe, expect, it } from "vitest";

import { nicknameError, rejectMessage } from "./nickname";

describe("nicknameError", () => {
  it.each(["abc", "Rocky_42", "abcdefghijklmnop"])("accepts %s", (n) => {
    expect(nicknameError(n)).toBeNull();
  });

  it.each([
    ["", "Digite um apelido."],
    ["ab", "O apelido precisa ter de 3 a 16 caracteres."],
    ["abcdefghijklmnopq", "O apelido precisa ter de 3 a 16 caracteres."],
    ["com espaço", "Use só letras, números e _."],
    ["rocky!", "Use só letras, números e _."],
  ])("rejects %j", (n, msg) => {
    expect(nicknameError(n)).toBe(msg);
  });
});

describe("rejectMessage", () => {
  it.each([
    ["nickname_taken", "Esse apelido já está em uso na Praça. Escolha outro."],
    ["nickname_invalid", "Apelido inválido. Use de 3 a 16 letras, números ou _."],
    ["room_full", "A Praça está cheia (10 guaxinins). Tente de novo daqui a pouco."],
    ["already_joined", "Você já está na Praça."],
  ] as const)("explains %s", (reason, msg) => {
    expect(rejectMessage(reason)).toBe(msg);
  });
});
