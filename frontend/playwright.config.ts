import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  timeout: 30_000,
  retries: 2,
  workers: 4,
  use: {
    trace: 'on-first-retry'
  }
  // No webServer — each test spawns its own lif-serve via fixtures
});
