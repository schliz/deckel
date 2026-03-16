import { test, expect } from '@playwright/test';

test.describe('Kiosk User Exclusion', () => {

  test('Kiosk-Anmeldung — kiosk user is redirected to /kiosk', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/kiosk/);
  });

  test('Kiosk-Layout — no header stats (balance/rank) shown', async ({ page }) => {
    await page.goto('/kiosk');

    // Kiosk layout should NOT contain header-stats
    await expect(page.locator('#header-stats')).not.toBeVisible();
    await expect(page.locator('text=Guthaben:')).not.toBeVisible();
    await expect(page.locator('text=Platz')).not.toBeVisible();
  });

  test('Kiosk-Menü — sees menu items for ordering', async ({ page }) => {
    await page.goto('/kiosk');

    // Should see item categories and items
    await expect(page.getByText('Getränke')).toBeVisible();
    await expect(page.getByText('Bier')).toBeVisible();
  });

  test('Kiosk-Benutzerauswahl — kiosk user not in target list', async ({ page }) => {
    await page.goto('/kiosk');

    // Click on an item to get to user selection
    const bierLink = page.locator('a', { hasText: 'Bier' });
    await bierLink.click();

    // Should see user selection page with regular users
    await expect(page.getByText('Test User', { exact: false })).toBeVisible({ timeout: 5000 });

    // Kiosk user itself should NOT appear in the list
    await expect(page.getByText('Test Kiosk', { exact: false })).not.toBeVisible();
  });

});
