import { expect } from "@playwright/test";

import { Then, When } from "./fixtures";

When("{string} envia a mensagem {string}", async ({ world }, nickname: string, text: string) => {
  const page = world.page(nickname);
  // The way a player chats: Enter opens the field, Enter sends.
  await page.keyboard.press("Enter");
  await expect(page.getByLabel("Mensagem")).toBeFocused();
  await page.keyboard.type(text);
  await page.keyboard.press("Enter");
});

// The balloon itself is drawn on the canvas; the accessible message log
// shows the same line and is what we can assert on.
Then(
  "{string} vê o balão {string} sobre {string}",
  async ({ world }, observer: string, text: string, sender: string) => {
    const log = world.page(observer).getByRole("list", { name: "Mensagens" });
    await expect(log).toContainText(`${sender}: ${text}`);
  },
);
