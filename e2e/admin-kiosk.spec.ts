import { test, expect } from '@playwright/test';

test.describe('Admin Kiosk User Management', () => {

  test('Kiosk-Benutzer in Liste — shows Kiosk badge and dashes for balance/actions', async ({ page }) => {
    await page.goto('/admin/users');

    const kioskRow = page.locator('tr', { hasText: 'testkiosk' });
    await expect(kioskRow).toBeVisible();

    // Should show "Kiosk" badge
    await expect(kioskRow.locator('.badge', { hasText: 'Kiosk' })).toBeVisible();

    // Balance column should show "—" (em dash), not a monetary value
    const balanceCell = kioskRow.locator('td').nth(2);
    await expect(balanceCell).toContainText('—');

    // Barteamer column should show "—"
    const barteamerCell = kioskRow.locator('td').nth(3);
    await expect(barteamerCell).toContainText('—');

    // Limit-Override column should show "—"
    const limitCell = kioskRow.locator('td').nth(5);
    await expect(limitCell).toContainText('—');

    // Actions column should show "—" (no deposit button)
    const actionsCell = kioskRow.locator('td').nth(6);
    await expect(actionsCell).toContainText('—');
  });

});
