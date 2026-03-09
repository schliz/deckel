import { test, expect } from '@playwright/test';

test.describe('Admin Stats & Settings', () => {
  test('Statistiken ansehen — admin opens stats page', async ({ page }) => {
    await page.goto('/admin/stats');

    // Wait for the stats panel to load (may be lazy-loaded via HTMX)
    const statsPanel = page.locator('#admin-stats-panel');
    await expect(statsPanel).toBeVisible();

    // Expect stat cards to be present
    await expect(statsPanel.getByText('Gesamtumsatz')).toBeVisible();
    await expect(statsPanel.getByText('Anzahl Transaktionen')).toBeVisible();
    await expect(statsPanel.getByText('Einzahlungen gesamt')).toBeVisible();

    // Expect category revenue section
    await expect(statsPanel.getByText('Umsatz nach Kategorie')).toBeVisible();

    // Expect top items sections to be present
    await expect(statsPanel.getByText('Meistverkaufte Artikel')).toBeVisible();
  });

  test('Zeitraum filtern — admin filters stats by preset and date range', async ({ page }) => {
    await page.goto('/admin/stats');

    const statsPanel = page.locator('#admin-stats-panel');
    await expect(statsPanel).toBeVisible();

    // Click the "today" preset button inside #stats-presets
    const presetsContainer = page.locator('#stats-presets');
    await expect(presetsContainer).toBeVisible();

    const todayBtn = presetsContainer.locator('button', { hasText: 'Heute' });
    await todayBtn.click();

    // Wait for the stats panel to update
    await expect(statsPanel).toBeVisible();
    await expect(statsPanel.getByText('Gesamtumsatz')).toBeVisible();

    // Test date range filtering with manual date inputs
    const fromInput = page.locator('input[name="from"]');
    const toInput = page.locator('input[name="to"]');
    await expect(fromInput).toBeVisible();
    await expect(toInput).toBeVisible();

    // Fill date inputs and submit
    await fromInput.fill('2025-01-01');
    await toInput.fill('2026-12-31');

    // Submit the date filter form — find the submit button near the date inputs
    const filterForm = fromInput.locator('xpath=ancestor::form');
    await filterForm.locator('button[type="submit"]').click();

    // Wait for stats panel to update with filtered data
    await expect(statsPanel).toBeVisible();
    await expect(statsPanel.getByText('Gesamtumsatz')).toBeVisible();
  });

  test('Einstellungen ansehen — admin views settings with correct seed values', async ({ page }) => {
    await page.goto('/admin/settings');

    // Verify limit fields exist and show correct seed values (cents converted to EUR)
    const warningLimit = page.locator('input[name="warning_limit"]');
    await expect(warningLimit).toBeVisible();
    await expect(warningLimit).toHaveValue('-10,00');

    const hardSpendingLimit = page.locator('input[name="hard_spending_limit"]');
    await expect(hardSpendingLimit).toBeVisible();
    await expect(hardSpendingLimit).toHaveValue('20,00');

    const hardLimitEnabled = page.locator('input[name="hard_limit_enabled"]');
    await expect(hardLimitEnabled).toBeVisible();
    await expect(hardLimitEnabled).toBeChecked();

    // Verify booking fields
    const customTxMin = page.locator('input[name="custom_tx_min"]');
    await expect(customTxMin).toBeVisible();
    await expect(customTxMin).toHaveValue('-5,00');

    const customTxMax = page.locator('input[name="custom_tx_max"]');
    await expect(customTxMax).toBeVisible();
    await expect(customTxMax).toHaveValue('5,00');

    const maxItemQuantity = page.locator('input[name="max_item_quantity"]');
    await expect(maxItemQuantity).toBeVisible();
    await expect(maxItemQuantity).toHaveValue('10');

    const cancellationMinutes = page.locator('input[name="cancellation_minutes"]');
    await expect(cancellationMinutes).toBeVisible();
    await expect(cancellationMinutes).toHaveValue('30');

    const paginationSize = page.locator('input[name="pagination_size"]');
    await expect(paginationSize).toBeVisible();
    await expect(paginationSize).toHaveValue('20');

    // Verify SMTP fields exist
    await expect(page.locator('input[name="smtp_host"]')).toBeVisible();
    await expect(page.locator('input[name="smtp_port"]')).toBeVisible();
  });

  test('Einstellungen speichern — admin changes and saves settings', async ({ page }) => {
    await page.goto('/admin/settings');

    // Change max_item_quantity to 15
    const maxItemQuantity = page.locator('input[name="max_item_quantity"]');
    await expect(maxItemQuantity).toBeVisible();
    await maxItemQuantity.clear();
    await maxItemQuantity.fill('15');

    // Submit the settings form
    const submitBtn = page.locator('form[hx-post="/admin/settings"] .btn.btn-primary');
    await submitBtn.click();

    // Expect a success toast in #toast-zone
    const toastZone = page.locator('#toast-zone');
    await expect(toastZone.locator('.alert.alert-success')).toBeVisible();

    // Reload and verify the value persisted
    await page.reload();
    await expect(page.locator('input[name="max_item_quantity"]')).toHaveValue('15');

    // Restore original value to not break other tests
    const restored = page.locator('input[name="max_item_quantity"]');
    await restored.clear();
    await restored.fill('10');
    await page.locator('form[hx-post="/admin/settings"] .btn.btn-primary').click();
    await expect(toastZone.locator('.alert.alert-success')).toBeVisible();

    // Verify restoration
    await page.reload();
    await expect(page.locator('input[name="max_item_quantity"]')).toHaveValue('10');
  });
});
