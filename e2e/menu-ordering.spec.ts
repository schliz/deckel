import { test, expect } from '@playwright/test';

test.describe('Menu & Ordering', () => {
  test('Karte ansehen — sees categories and correct Helfer prices', async ({ page }) => {
    await page.goto('/');

    // Expect both categories visible
    await expect(page.getByText('Getränke')).toBeVisible();
    await expect(page.getByText('Snacks')).toBeVisible();

    // Expect items with correct Helfer prices
    const bierRow = page.locator('li.list-row', { hasText: 'Bier' });
    await expect(bierRow).toBeVisible();
    await expect(bierRow).toContainText('2,00');

    const mateRow = page.locator('li.list-row', { hasText: 'Mate' });
    await expect(mateRow).toBeVisible();
    await expect(mateRow).toContainText('2,00');

    const chipsRow = page.locator('li.list-row', { hasText: 'Chips' });
    await expect(chipsRow).toBeVisible();
    await expect(chipsRow).toContainText('1,50');
  });

  test('Getränk bestellen — order an item and see success', async ({ page }) => {
    await page.goto('/');

    // Click on the Bier menu item to open order modal
    const bierRow = page.locator('li.list-row', { hasText: 'Bier' });
    await bierRow.click();

    // Wait for modal to become visible
    const modal = page.locator('#modal');
    await expect(modal.locator('.modal-open')).toBeVisible();

    // Verify quantity input defaults to 1
    const qtyInput = modal.locator('#order-qty');
    await expect(qtyInput).toHaveValue('1');

    // Confirm the order
    const confirmBtn = modal.locator('#order-confirm-btn');
    await expect(confirmBtn).toBeVisible();
    await confirmBtn.click();

    // Expect success modal (still open with OK button)
    await expect(modal.locator('.modal-open')).toBeVisible();

    // Click OK to dismiss
    const okBtn = modal.locator('.btn.btn-outline.btn-lg');
    await expect(okBtn).toBeVisible();
    await okBtn.click();
  });

  test('Guthabenanzeige — header shows balance, total, and rank', async ({ page }) => {
    await page.goto('/');

    // Header stats load lazily via hx-get="/header-stats"
    const headerStats = page.locator('#header-stats');
    await expect(headerStats).toContainText('Guthaben:', { timeout: 10000 });
    await expect(headerStats).toContainText('Saldo Gesamtliste:');
    await expect(headerStats).toContainText('Platz');
  });

  test('Navigation — user sees Karte, Transaktionen, Profil but no admin links', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('link', { name: 'Karte' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Transaktionen' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Profil' })).toBeVisible();

    // No admin menu should be visible
    await expect(page.getByText('Admin:')).not.toBeVisible();
  });

  test('Themenumschaltung — theme toggle switches data-theme', async ({ page }) => {
    await page.goto('/');

    const html = page.locator('html');
    const initialTheme = await html.getAttribute('data-theme');

    // Click the theme toggle button
    const themeToggle = page.locator('[onclick="toggleTheme()"]');
    await themeToggle.click();

    const newTheme = await html.getAttribute('data-theme');
    expect(newTheme).not.toBe(initialTheme);

    // Verify it's one of the two expected themes
    expect(['corporate2', 'business2']).toContain(newTheme);

    // Toggle back
    await themeToggle.click();
    const restoredTheme = await html.getAttribute('data-theme');
    expect(restoredTheme).toBe(initialTheme);
  });
});
