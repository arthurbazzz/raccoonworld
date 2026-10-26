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

// Visual assertions read data-facing / data-moving / data-ring from the
// accessible player list; they land in RW-002/T7 and T8.
// playwright-bdd checks each function's arity against its step pattern and
// needs the fixtures argument destructured.
const PENDING = "pending: RW-002/T8";
/* eslint-disable @typescript-eslint/no-unused-vars */
const pending1 = async ({ world }: { world: unknown }, _a: string) => {
  throw new Error(PENDING);
};
const pending2 = async ({ world }: { world: unknown }, _a: string, _b: string) => {
  throw new Error(PENDING);
};
const pending3 = async ({ world }: { world: unknown }, _a: string, _b: string, _c: string) => {
  throw new Error(PENDING);
};
/* eslint-enable @typescript-eslint/no-unused-vars */

Then(/^"([^"]*)" vê "([^"]*)" olhando para (cima|baixo|a esquerda|a direita)$/, pending3);
Then(/^"([^"]*)" está parado olhando para (cima|baixo|a esquerda|a direita)$/, pending2);
Then("{string} vê {string} andando", pending2);
Then("{string} vê {string} parado", pending2);
Then("{string} vê o anel embaixo do próprio guaxinim", pending1);
Then("{string} não vê anel embaixo do guaxinim de {string}", pending2);
