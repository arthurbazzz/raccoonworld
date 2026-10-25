import { expect } from "@playwright/test";

import { Given, Then, When } from "./fixtures";

/** How long "segura a tecla" holds before the next step checks movement. */
const HOLD_MS = 500;
/** Longer than a server tick plus network, so a wrong move would show up. */
const SETTLE_MS = 400;

When("{string} segura a tecla {string}", async ({ world }, nickname: string, key: string) => {
  await world.mark(nickname);
  await world.keyDown(nickname, key.toLowerCase());
  await world.page(nickname).waitForTimeout(HOLD_MS);
});

Then("{string} se move para a direita", async ({ world }, nickname: string) => {
  const start = world.marks.get(nickname)!;
  await expect
    .poll(async () => (await world.position(nickname, nickname)).x)
    .toBeGreaterThan(start.x);
  expect((await world.position(nickname, nickname)).y).toBe(start.y);
});

Then(
  "{string} vê {string} se mover para a direita",
  async ({ world }, observer: string, mover: string) => {
    const start = world.marks.get(mover)!;
    await expect
      .poll(async () => (await world.position(observer, mover)).x)
      .toBeGreaterThan(start.x);
  },
);

Given("que {string} está com o campo de chat em foco", async ({ world }, nickname: string) => {
  if (!world.nicknames().includes(nickname)) await world.join(nickname);
  const page = world.page(nickname);
  await world.mark(nickname);
  const chat = page.getByLabel("Mensagem");
  await chat.click();
  await expect(chat).toBeFocused();
});

// Each key is held long enough that the server would move the raccoon if
// the keystroke leaked into movement (RN3.1).
When("{string} digita {string}", async ({ world }, nickname: string, text: string) => {
  const page = world.page(nickname);
  for (const ch of text) {
    await page.keyboard.down(ch);
    await page.waitForTimeout(150);
    await page.keyboard.up(ch);
  }
  await expect(page.getByLabel("Mensagem")).toHaveValue(text);
});

Then("a posição de {string} não muda", async ({ world }, nickname: string) => {
  await world.page(nickname).waitForTimeout(SETTLE_MS);
  expect(await world.position(nickname, nickname)).toEqual(world.marks.get(nickname));
});
