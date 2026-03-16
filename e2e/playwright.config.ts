import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  fullyParallel: false,
  retries: 0,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:4181',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'setup', testMatch: /.*\.setup\.ts/ },
    {
      name: 'user-tests',
      testIgnore: /admin-|kiosk-/,
      use: {
        ...devices['Desktop Chrome'],
        storageState: '.auth/user.json',
      },
      dependencies: ['setup'],
    },
    {
      name: 'admin-tests',
      testMatch: /admin-/,
      use: {
        ...devices['Desktop Chrome'],
        storageState: '.auth/admin.json',
      },
      dependencies: ['setup'],
    },
    {
      name: 'kiosk-tests',
      testMatch: /kiosk-/,
      use: {
        ...devices['Desktop Chrome'],
        storageState: '.auth/kiosk.json',
      },
      dependencies: ['setup'],
    },
  ],
});
