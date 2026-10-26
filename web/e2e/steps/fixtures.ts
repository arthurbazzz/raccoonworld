import { expect, type Browser, type BrowserContext, type Page } from "@playwright/test";
import { createBdd, test as base } from "playwright-bdd";

export const SERVER_URL = "http://localhost:8080";

interface Player {
  context: BrowserContext;
  page: Page;
}

export interface Point {
  x: number;
  y: number;
}

/**
 * Per-scenario state: one isolated browser context per player, like two
 * people on two computers. Closing the contexts disconnects everyone, so
 * the next scenario starts with an empty Praça.
 */
export class World {
  private readonly players = new Map<string, Player>();
  private readonly held = new Map<string, Set<string>>();
  readonly marks = new Map<string, Point>();

  constructor(private readonly browser: Browser) {}

  async join(nickname: string): Promise<Page> {
    const context = await this.browser.newContext();
    const page = await context.newPage();
    this.players.set(nickname, { context, page });

    await page.goto("/");
    await page.getByLabel("Apelido").fill(nickname);
    await page.getByRole("button", { name: "Entrar na Praça" }).click();
    await expect(this.entry(page, nickname)).toBeAttached();
    return page;
  }

  page(nickname: string): Page {
    const p = this.players.get(nickname);
    if (!p) throw new Error(`${nickname} never joined`);
    return p.page;
  }

  nicknames(): string[] {
    return [...this.players.keys()];
  }

  /** The player list entry for `nickname` as seen from `page`. */
  entry(page: Page, nickname: string) {
    return page
      .getByRole("list", { name: "Jogadores na Praça" })
      .locator(`li[data-player="${nickname}"]`);
  }

  /** Where `observer`'s client believes `target` is (server position). */
  async position(observer: string, target: string): Promise<Point> {
    const li = this.entry(this.page(observer), target);
    const [x, y] = await Promise.all([li.getAttribute("data-x"), li.getAttribute("data-y")]);
    return { x: Number(x), y: Number(y) };
  }

  async mark(nickname: string): Promise<void> {
    this.marks.set(nickname, await this.position(nickname, nickname));
  }

  async keyDown(nickname: string, key: string): Promise<void> {
    await this.page(nickname).keyboard.down(key);
    const keys = this.held.get(nickname) ?? new Set<string>();
    keys.add(key);
    this.held.set(nickname, keys);
  }

  async keyUp(nickname: string, key: string): Promise<void> {
    await this.page(nickname).keyboard.up(key);
    this.held.get(nickname)?.delete(key);
  }

  async close(): Promise<void> {
    for (const [nickname, keys] of this.held) {
      for (const key of keys) await this.page(nickname).keyboard.up(key);
    }
    await Promise.all([...this.players.values()].map((p) => p.context.close()));
  }
}

export const test = base.extend<{ world: World }>({
  // Named `provide`, not `use`, so the React hooks lint rule leaves it alone.
  world: async ({ browser }, provide) => {
    const world = new World(browser);
    await provide(world);
    await world.close();
  },
});

export const { Given, When, Then } = createBdd(test);
