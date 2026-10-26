import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

// Unit tests only. Browser BDD scenarios run with playwright-bdd (e2e/).
export default defineConfig({
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  test: {
    include: ["src/**/*.test.ts", "scripts/**/*.test.ts"],
    environment: "node",
  },
});
