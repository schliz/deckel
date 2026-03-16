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

// Helper: toggle spending limit exemption for a user via admin page.
async function toggleSpendingLimitExemption(ctx: BrowserContext, baseURL: string, userName: string) {
  const page = await ctx.newPage();
  await page.goto(`${baseURL}/admin/users`);
  const row = page.locator('tr', { hasText: userName });
  // The spending limit toggle is in the 6th column (index 5)
  const toggle = row.locator('td').nth(5).locator('input[type="checkbox"]');
  await toggle.click();
  // Confirm in modal
  const modal = page.locator('#modal');
  await modal.locator('button', { hasText: 'Bestätigen' }).click();
  // Wait for the row to update
  await expect(row).toBeVisible();
  await page.close();
}

test.describe('Spending Limits', () => {

  // Note: These tests run after menu-ordering.spec.ts (alphabetical order).
  // The test user already has a -2 EUR balance from ordering one Bier.

  test('Bestellung bei Ausgabelimit — menu blocked when at limit', async ({ page, browser }) => {
    const baseURL = 'http://localhost:4181';
    const admin = await adminContext(browser);

    // Set a low hard limit (1 EUR). User already has -2 EUR from menu-ordering,
    // so they are already at/below -1 EUR and should be blocked immediately.
    await setSpendingLimit(admin, baseURL, '1', true);

    // Menu should show blocked state
    await page.goto('/');
    await expect(page.locator('.alert-error')).toBeVisible();
    await expect(page.locator('.alert-error')).toContainText('Ausgabelimit');

    // Items should show "Bitte erst einzahlen" badge instead of price
    const bierRow = page.locator('li.list-row', { hasText: 'Bier' });
    await expect(bierRow.locator('.badge-warning')).toContainText('Bitte erst einzahlen');

    // Eigene Buchung should also be blocked
    const customRow = page.locator('li.list-row', { hasText: 'Eigene Buchung' });
    await expect(customRow).toHaveClass(/opacity-50/);

    // Restore settings
    await setSpendingLimit(admin, baseURL, '20', true);
    await admin.close();
  });

  test('Bestellung bei deaktiviertem Limit — exempt user can order despite limit', async ({ page, browser }) => {
    const baseURL = 'http://localhost:4181';
    const admin = await adminContext(browser);

    // Set a low hard limit — user is at -2 EUR, limit=1 EUR → blocked
    await setSpendingLimit(admin, baseURL, '1', true);

    // Exempt the test user from spending limits
    await toggleSpendingLimitExemption(admin, baseURL, 'Test User');

    // User should be able to order despite negative balance
    await page.goto('/');
    // Should NOT see blocked alert
    await expect(page.locator('.alert-error')).not.toBeVisible();

    // Order should succeed
    const bierRow = page.locator('li.list-row', { hasText: 'Bier' });
    await bierRow.click();

    const modal = page.locator('#modal');
    await expect(modal.locator('.modal-open')).toBeVisible();
    await modal.locator('#order-confirm-btn').click();

    // Should see success overlay
    await expect(modal.locator('.modal-open')).toBeVisible();
    await modal.locator('.btn.btn-outline.btn-lg').click();

    // Restore: re-enable limit for user, restore settings
    await toggleSpendingLimitExemption(admin, baseURL, 'Test User');
    await setSpendingLimit(admin, baseURL, '20', true);
    await admin.close();
  });

  test('Bestellung bei niedrigem Guthaben — warning shown after order', async ({ page, browser }) => {
    const baseURL = 'http://localhost:4181';
    const admin = await adminContext(browser);

    // Set warning limit to 0 EUR — any negative balance triggers warning.
    // Set hard limit high enough to not block (user has about -4 EUR at this point).
    await setWarningLimit(admin, baseURL, '0');
    await setSpendingLimit(admin, baseURL, '100', true);

    // Order — user has negative balance, so warning should appear
    await page.goto('/');
    const bierRow = page.locator('li.list-row', { hasText: 'Bier' });
    await bierRow.click();

    const modal = page.locator('#modal');
    await expect(modal.locator('.modal-open')).toBeVisible();
    await modal.locator('#order-confirm-btn').click();

    // The success overlay should show warning text (amber background + warning message)
    await expect(page.getByText('bald einzahlen')).toBeVisible();

    // Dismiss
    await modal.locator('.btn.btn-outline.btn-lg').click();

    // Restore warning limit and spending limit
    await setWarningLimit(admin, baseURL, '-10');
    await setSpendingLimit(admin, baseURL, '20', true);
    await admin.close();
  });

});
