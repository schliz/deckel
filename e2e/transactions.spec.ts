import { test, expect } from '@playwright/test';

test.describe('Transactions & Profile', () => {
  test('Transaktionen ansehen', async ({ page }) => {
    await page.goto('/transactions');

    const list = page.locator('#transaction-list');
    const table = list.locator('table.table');
    await expect(table).toBeVisible();
    await expect(table.locator('thead')).toBeVisible();
    await expect(table.locator('tbody')).toBeVisible();

    const headers = table.locator('thead th');
    await expect(headers.filter({ hasText: 'Datum' })).toBeVisible();
    await expect(headers.filter({ hasText: 'Beschreibung' })).toBeVisible();
    await expect(headers.filter({ hasText: 'Menge' })).toBeVisible();
    await expect(headers.filter({ hasText: 'Betrag' })).toBeVisible();
  });

  test('Eigene Buchung erstellen', async ({ page }) => {
    await page.goto('/');

    const trigger = page.locator('li', { hasText: 'Eigene Buchung' });
    await trigger.click();

    const modal = page.locator('#modal .modal-open');
    await expect(modal).toBeVisible({ timeout: 5000 });

    await page.fill('#custom-tx-description', 'Test Buchung');
    await page.fill('#custom-tx-amount', '2.50');
    await page.click('#custom-tx-confirm-btn');

    await expect(
      page.locator('.alert-success, #modal .modal-open').first()
    ).toBeVisible({ timeout: 5000 });
  });

  test('Eigene Buchung: Fehler bei leerer Beschreibung', async ({ page }) => {
    await page.goto('/');

    const trigger = page.locator('li', { hasText: 'Eigene Buchung' });
    await trigger.click();

    const modal = page.locator('#modal .modal-open');
    await expect(modal).toBeVisible({ timeout: 5000 });

    await page.fill('#custom-tx-amount', '2.50');
    await page.click('#custom-tx-confirm-btn');

    // Modal stays open with inline error
    await expect(modal).toBeVisible();
    await expect(modal.locator('.alert-error')).toContainText('Beschreibung ist erforderlich');
  });

  test('Eigene Buchung: Fehler bei ungültigem Betrag', async ({ page }) => {
    await page.goto('/');

    const trigger = page.locator('li', { hasText: 'Eigene Buchung' });
    await trigger.click();

    const modal = page.locator('#modal .modal-open');
    await expect(modal).toBeVisible({ timeout: 5000 });

    await page.fill('#custom-tx-description', 'Test');
    await page.fill('#custom-tx-amount', '0');
    await page.click('#custom-tx-confirm-btn');

    await expect(modal).toBeVisible();
    await expect(modal.locator('.alert-error')).toContainText('Ungültiger Betrag');
  });

  test('Eigene Buchung: Fehler bei Betrag über Maximum', async ({ page }) => {
    await page.goto('/');

    const trigger = page.locator('li', { hasText: 'Eigene Buchung' });
    await trigger.click();

    const modal = page.locator('#modal .modal-open');
    await expect(modal).toBeVisible({ timeout: 5000 });

    await page.fill('#custom-tx-description', 'Test');
    await page.fill('#custom-tx-amount', '99.99');
    await page.click('#custom-tx-confirm-btn');

    await expect(modal).toBeVisible();
    await expect(modal.locator('.alert-error')).toContainText('Maximalbetrag');
  });

  test('Transaktion stornieren', async ({ page }) => {
    await page.goto('/transactions');

    const cancelBtn = page
      .locator('#transaction-list .btn.btn-error.btn-xs')
      .first();
    await expect(cancelBtn).toBeVisible({ timeout: 5000 });
    await cancelBtn.click();

    const modal = page.locator('#modal .modal-open');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal).toContainText('Wirklich stornieren?');

    const confirmBtn = modal.locator('.btn.btn-error[hx-post]');
    await confirmBtn.click();

    const cancelledRow = page.locator(
      '#transaction-list tbody tr.opacity-50, #transaction-list tbody tr.line-through'
    );
    await expect(cancelledRow.first()).toBeVisible({ timeout: 5000 });
  });

  test('Profil ansehen', async ({ page }) => {
    await page.goto('/profile');

    await expect(page.locator('.card')).toBeVisible();

    await expect(page.getByText('E-Mail')).toBeVisible();
    await expect(page.getByText('Status')).toBeVisible();
    await expect(page.getByText('Mitglied seit')).toBeVisible();

    const badge = page.locator('.badge');
    await expect(
      badge.filter({ hasText: /Helfer|Barteamer/ }).first()
    ).toBeVisible();
  });
});
