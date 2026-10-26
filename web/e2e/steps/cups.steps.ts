import { expect, type Page } from "@playwright/test";

import { type CupsRound, Given, Then, When, type World } from "./fixtures";

// Step texts match specs/005-tres-copos/features/tres_copos.feature verbatim.
// The server runs with RW_CUPS_FAST=1 (playwright.config.ts), so a round
// takes about a second.

/** Walking legs from the spawn point around the fountain to the booth. */
const LEGS: { key: string; done: (p: { x: number; y: number }) => boolean }[] = [
  { key: "a", done: (p) => p.x <= 420 },
  { key: "w", done: (p) => p.y <= 200 },
  { key: "d", done: (p) => p.x >= 480 },
];

async function walkToBooth(world: World, nickname: string): Promise<void> {
  for (const leg of LEGS) {
    await world.keyDown(nickname, leg.key);
    await expect
      .poll(async () => leg.done(await world.position(nickname, nickname)), { timeout: 10_000 })
      .toBe(true);
    await world.keyUp(nickname, leg.key);
  }
}

const panel = (page: Page) => page.getByRole("region", { name: "Três copos" });

/** Where the nut ends up in the last round `nickname` received. */
function nut(world: World, nickname: string): number {
  const rounds = world.rounds.get(nickname) ?? [];
  const r: CupsRound | undefined = rounds[rounds.length - 1];
  if (!r) throw new Error(`${nickname} received no round`);
  let slot = r.nut;
  for (const [a, b] of r.swaps) slot = slot === a ? b : slot === b ? a : slot;
  return slot;
}

async function join(world: World, nickname: string): Promise<void> {
  if (!world.nicknames().includes(nickname)) await world.join(nickname);
}

async function startPlaying(world: World, nicknames: string[]): Promise<void> {
  for (const n of nicknames) {
    await join(world, n);
    await walkToBooth(world, n);
  }
  // Everyone walks first: the fast lobby lasts a second.
  for (const n of nicknames) await world.page(n).keyboard.press("e");
  for (const n of nicknames) {
    await expect(panel(world.page(n))).toHaveAttribute("data-round", "1");
  }
}

async function pick(world: World, nickname: string, slot: number): Promise<void> {
  await panel(world.page(nickname)).locator(`button[data-slot="${slot}"]`).click();
}

Given("que a barraca dos três copos está livre", async () => {
  // Players of earlier scenarios left with their browser contexts, which
  // frees the booth (RC14).
});

Given("que {string} está na Praça longe da barraca", async ({ world }, nickname: string) => {
  await join(world, nickname);
  await expect(world.page(nickname).getByText("Aperte E para jogar")).toBeHidden();
});

When("{string} anda até perto da barraca", async ({ world }, nickname: string) => {
  await walkToBooth(world, nickname);
});

Then("{string} vê o aviso {string}", async ({ world }, nickname: string, text: string) => {
  await expect(world.page(nickname).getByText(text)).toBeVisible();
});

Given(
  "que {string} e {string} estão numa rodada com o embaralhamento terminado",
  async ({ world }, a: string, b: string) => {
    await startPlaying(world, [a, b]);
    for (const n of [a, b]) {
      await expect(panel(world.page(n))).toHaveAttribute("data-stage", "choose");
    }
  },
);

When("{string} escolhe o copo onde está a noz", async ({ world }, nickname: string) => {
  await pick(world, nickname, nut(world, nickname));
});

When("{string} escolhe um copo sem a noz", async ({ world }, nickname: string) => {
  await pick(world, nickname, (nut(world, nickname) + 1) % 3);
});

Then("{string} vê {string}", async ({ world }, nickname: string, text: string) => {
  await expect(panel(world.page(nickname)).getByText(text)).toBeVisible();
});

Then("os dois veem que só {string} acertou a rodada", async ({ world }, nickname: string) => {
  for (const n of world.nicknames()) {
    await expect(panel(world.page(n))).toContainText(`Acertaram: ${nickname}`);
  }
});

Given("que {string} está jogando os três copos", async ({ world }, nickname: string) => {
  await startPlaying(world, [nickname]);
});

Given("que {string} está na Praça sem jogar", async ({ world }, nickname: string) => {
  await join(world, nickname);
});

When("os copos são embaralhados", async ({ world }) => {
  const player = world.nicknames()[0];
  await expect(panel(world.page(player))).toHaveAttribute("data-stage", "shuffle");
});

Then("{string} vê os copos se mexendo na barraca", async ({ world }, nickname: string) => {
  const page = world.page(nickname);
  await expect(panel(page)).toBeHidden(); // not playing
  await expect(page.getByLabel("Barraca dos três copos")).toHaveAttribute("data-stage", "shuffle");
});

When("{string} começa uma partida", async ({ world }, nickname: string) => {
  await walkToBooth(world, nickname);
  await world.page(nickname).keyboard.press("e");
});

Then(
  "{string} vê o sinal de jogando sobre {string}",
  async ({ world }, observer: string, player: string) => {
    await expect(world.entry(world.page(observer), player)).toHaveAttribute("data-cups", "player");
  },
);
