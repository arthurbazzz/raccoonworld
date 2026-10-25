import { defineConfig, devices } from "@playwright/test";
import { defineBddConfig } from "playwright-bdd";

// Browser BDD (constitution, principle IV): the .feature files in specs/ are
// the single source; only scenarios tagged @ui or @e2e run here. Every
// scenario shares one real server, so they run one at a time.
const testDir = defineBddConfig({
  features: "../specs/*/features/*.feature",
  featuresRoot: "../specs",
  steps: "e2e/steps/*.ts",
  tags: "@ui or @e2e",
});

export default defineConfig({
  testDir,
  fullyParallel: false,
  workers: 1,
  timeout: 30_000,
  reporter: [["list"]],
  use: {
    baseURL: "http://localhost:3000",
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: [
    {
      command: "go run ./cmd/server",
      cwd: "../server",
      url: "http://localhost:8080/health",
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      command: "npm run dev",
      url: "http://localhost:3000",
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
  ],
});
