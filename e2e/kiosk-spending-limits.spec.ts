import { test, expect, BrowserContext } from '@playwright/test';
import path from 'path';

// Helper: create an admin browser context for setup operations.
async function adminContext(browser: any): Promise<BrowserContext> {
  return browser.newContext({
    storageState: path.join(__dirname, '.auth/admin.json'),
  });
}

// Helper: update spending limit settings via admin page.
async function setSpendingLimit(ctx: BrowserContext, baseURL: string, limitEur: string, enabled: boolean) {
  const page = await ctx.newPage();
  await page.goto(`${baseURL}/admin/settings`);
  await page.locator('input[name="hard_spending_limit"]').fill(limitEur);
  const checkbox = page.locator('input[name="hard_limit_enabled"]');
  if (enabled) {
    await checkbox.check();
  } else {
    await checkbox.uncheck();
  }
  await page.locator('button[type="submit"]').click();
  await expect(page.locator('.alert-success')).toBeVisible();
  await page.close();
}

// Helper: update warning limit via admin page.
async function setWarningLimit(ctx: BrowserContext, baseURL: string, limitEur: string) {
  const page = await ctx.newPage();
  await page.goto(`${baseURL}/admin/settings`);
  await page.locator('input[name="warning_limit"]').fill(limitEur);
  await page.locator('button[type="submit"]').click();
  await expect(page.locator('.alert-success')).toBeVisible();
  await page.close();
}

// Helper: place a kiosk order for a user. Returns after the success overlay is shown.
async function kioskOrder(page: any, itemName: string, userName: string, quantity: number) {
  await page.goto('/kiosk');
  await page.locator('a.card', { hasText: itemName }).click();
  await page.locator('.user-card', { hasText: userName }).click();

  // Adjust quantity if needed
  if (quantity > 1) {
    for (let i = 1; i < quantity; i++) {
      await page.locator('#qty-plus').click();
    }
  }

  await page.locator('#confirm-btn').click();
  // Wait for success overlay
  await expect(page.getByText('Bestellung gebucht!')).toBeVisible();
  // Wait for redirect back to kiosk menu
  await page.waitForURL('**/kiosk');
}

test.describe('Kiosk Spending Limits', () => {

  // Note: These tests run after user-tests and admin-tests.
  // The admin-users test deposits 10 EUR for testuser, so the balance
  // is positive at this point. We drain it via kiosk orders first.

  test('Kiosk-Bestellung bei Ausgabelimit — blocked user cannot order via kiosk', async ({ page, browser }) => {
    const baseURL = 'http://localhost:4181';
    const admin = await adminContext(browser);

    // Ensure hard limit is disabled first so we can place orders freely
    await setSpendingLimit(admin, baseURL, '20', false);

    // Place kiosk orders to drain Test User's balance into negative territory.
    // Order 5x Bier (5 * 2 EUR = 10 EUR) to ensure balance goes well below zero.
    await kioskOrder(page, 'Bier', 'Test User', 5);

    // Now enable a low hard limit (1 EUR). User should be at/below -1 EUR.
    await setSpendingLimit(admin, baseURL, '1', true);

    // Navigate to kiosk, select an item
    await page.goto('/kiosk');
    await page.locator('a.card', { hasText: 'Bier' }).click();

    // On user selection page, Test User should show "Limit erreicht" badge
    const userCard = page.locator('.user-card', { hasText: 'Test User' });
    await expect(userCard.locator('.badge')).toContainText('Limit erreicht');

    // Click on blocked user
    await userCard.click();

    // Confirm page should show error alert and disabled button
    await expect(page.locator('.alert-error')).toBeVisible();
    await expect(page.locator('.alert-error')).toContainText('Ausgabelimit');
    await expect(page.locator('#confirm-btn')).toBeDisabled();

    // Restore settings
    await setSpendingLimit(admin, baseURL, '20', true);
    await admin.close();
  });

  test('Kiosk-Bestellung bei niedrigem Guthaben — warning shown on confirm', async ({ page, browser }) => {
    const baseURL = 'http://localhost:4181';
    const admin = await adminContext(browser);

    // Set warning limit to 0 so any negative balance triggers warning.
    // Set hard limit high enough to not block.
    await setWarningLimit(admin, baseURL, '0');
    await setSpendingLimit(admin, baseURL, '100', true);

    // Navigate kiosk flow: select item, then user
    await page.goto('/kiosk');
    await page.locator('a.card', { hasText: 'Bier' }).click();
    await page.locator('.user-card', { hasText: 'Test User' }).click();

    // Should see low balance warning on confirm page (user has negative balance)
    await expect(page.locator('.alert-warning')).toBeVisible();
    await expect(page.locator('.alert-warning')).toContainText('Niedriges Guthaben');

    // Restore settings
    await setWarningLimit(admin, baseURL, '-10');
    await setSpendingLimit(admin, baseURL, '20', true);
    await admin.close();
  });

});
