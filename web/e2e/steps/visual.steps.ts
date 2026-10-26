import { expect } from "@playwright/test";

import { Given, Then, When } from "./fixtures";

// Steps for specs/002-visual-guaxinim/features/visual.feature.

/** How long a key change is held before the next step looks at the result. */
const HOLD_MS = 500;

When(
  "{string} segura as teclas {string} e {string}",
  async ({ world }, nickname: string, k1: string, k2: string) => {
    await world.mark(nickname);
    await world.keyDown(nickname, k1.toLowerCase());
    await world.keyDown(nickname, k2.toLowerCase());
    await world.page(nickname).waitForTimeout(HOLD_MS);
  },
);

Given(
  "que {string} está segurando a tecla {string}",
  async ({ world }, nickname: string, key: string) => {
    if (!world.nicknames().includes(nickname)) await world.join(nickname);
    await world.keyDown(nickname, key.toLowerCase());
    await world.page(nickname).waitForTimeout(HOLD_MS);
  },
);

When("{string} solta a tecla {string}", async ({ world }, nickname: string, key: string) => {
  await world.keyUp(nickname, key.toLowerCase());
  await world.page(nickname).waitForTimeout(HOLD_MS);
});

// Visual assertions read the accessible player list, where the game exposes
// what the scene draws: data-facing, data-moving and data-ring (RW-002/T7).
// toHaveAttribute retries, which covers the 100ms refresh and the 100ms
// interpolation delay.
//
// Facing checks retry for at most FACING_TIMEOUT_MS. With the default 5s, a
// raccoon still holding a diagonal reaches a wall and slides straight along
// it, so a wrong diagonal rule would eventually show the "right" answer and
// the test would pass (found by mutation in RW-002/T8).
const FACING_TIMEOUT_MS = 1000;

const FACING: Record<string, string> = {
  cima: "up",
  baixo: "down",
  "a esquerda": "left",
  "a direita": "right",
};

Then(
  /^"([^"]*)" vê "([^"]*)" olhando para (cima|baixo|a esquerda|a direita)$/,
  async ({ world }, observer: string, target: string, direction: string) => {
    const entry = world.entry(world.page(observer), target);
    await expect(entry).toHaveAttribute("data-facing", FACING[direction], {
      timeout: FACING_TIMEOUT_MS,
    });
  },
);

Then(
  /^"([^"]*)" está parado olhando para (cima|baixo|a esquerda|a direita)$/,
  async ({ world }, nickname: string, direction: string) => {
    const entry = world.entry(world.page(nickname), nickname);
    await expect(entry).toHaveAttribute("data-moving", "false");
    await expect(entry).toHaveAttribute("data-facing", FACING[direction], {
      timeout: FACING_TIMEOUT_MS,
    });
  },
);

Then("{string} vê {string} andando", async ({ world }, observer: string, target: string) => {
  await expect(world.entry(world.page(observer), target)).toHaveAttribute("data-moving", "true");
});

Then("{string} vê {string} parado", async ({ world }, observer: string, target: string) => {
  await expect(world.entry(world.page(observer), target)).toHaveAttribute("data-moving", "false");
});

Then("{string} vê o anel embaixo do próprio guaxinim", async ({ world }, nickname: string) => {
  await expect(world.entry(world.page(nickname), nickname)).toHaveAttribute("data-ring", "true");
});

Then(
  "{string} não vê anel embaixo do guaxinim de {string}",
  async ({ world }, observer: string, other: string) => {
    await expect(world.entry(world.page(observer), other)).toHaveAttribute("data-ring", "false");
  },
);
