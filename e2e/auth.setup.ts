import { test as setup } from '@playwright/test';
import path from 'path';

const userFile = path.join(__dirname, '.auth/user.json');
const adminFile = path.join(__dirname, '.auth/admin.json');
const kioskFile = path.join(__dirname, '.auth/kiosk.json');

setup('authenticate as regular user', async ({ page }) => {
  await page.goto('/');
  await page.locator('#username').fill('testuser');
  await page.locator('#password').fill('testpass');
  await page.locator('#kc-login').click();
  await page.waitForURL('**/');
  await page.context().storageState({ path: userFile });
});

setup('authenticate as admin', async ({ page }) => {
  await page.goto('/');
  await page.locator('#username').fill('testadmin');
  await page.locator('#password').fill('testpass');
  await page.locator('#kc-login').click();
  await page.waitForURL('**/');
  await page.context().storageState({ path: adminFile });
});

setup('authenticate as kiosk', async ({ page }) => {
  await page.goto('/');
  await page.locator('#username').fill('testkiosk');
  await page.locator('#password').fill('testpass');
  await page.locator('#kc-login').click();
  await page.waitForURL('**/kiosk');
  await page.context().storageState({ path: kioskFile });
});
