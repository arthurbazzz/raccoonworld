import { expect } from "@playwright/test";

import { Given, SERVER_URL, Then, When } from "./fixtures";

// Step texts match specs/001-sala-inicial/features/*.feature verbatim.

Given("que a sala {string} está aberta", async ({ request }, name: string) => {
  expect(name).toBe("Praça");
  expect((await request.get(`${SERVER_URL}/health`)).ok()).toBe(true);
});

When("o visitante entra com o apelido {string}", async ({ world }, nickname: string) => {
  await world.join(nickname);
});

Given("que {string} está na Praça", async ({ world }, nickname: string) => {
  await world.join(nickname);
});

Given("que {string} e {string} estão na Praça", async ({ world }, a: string, b: string) => {
  await world.join(a);
  await world.join(b);
});

Then("{string} aparece na Praça", async ({ world }, nickname: string) => {
  await expect(world.entry(world.page(nickname), nickname)).toBeAttached();
});

Then("{string} recebe a lista dos jogadores já presentes", async ({ world }, nickname: string) => {
  const page = world.page(nickname);
  for (const other of world.nicknames()) {
    await expect(world.entry(page, other)).toBeAttached();
  }
});
