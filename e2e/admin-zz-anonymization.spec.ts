import { test, expect } from '@playwright/test';

// Anonymization + user-deletion (see docs/anonymisierung.md and test/seed.sql).
//
// This spec permanently mutates shared DB state (it runs the real
// anonymization job against the fixtures seeded for user 901 "Old Anon User"
// and deletes user 902 "Delete Me"), so it is named to sort alphabetically
// after every other `admin-*.spec.ts` file and must only run once per test
// database. It must run in the admin-tests project (storage state:
// .auth/admin.json → testadmin@example.com).

const ADMIN_EMAIL = 'testadmin@example.com';

test.describe('Admin Anonymisierung (DSGVO)', () => {
  test('Einstellungsseite zeigt Anonymisierungs-Karte mit Frist 36 Monate', async ({ page }) => {
    await page.goto('/admin/settings');

    const anonymizeCard = page.locator('.card', { hasText: 'Anonymisierung (DSGVO)' });
    await expect(anonymizeCard).toBeVisible();

    const monthsInput = page.locator('input[name="anonymize_after_months"]');
    await expect(monthsInput).toBeVisible();
    await expect(monthsInput).toHaveValue('36');

    await expect(anonymizeCard.locator('button', { hasText: 'Jetzt anonymisieren' })).toBeVisible();
  });

  test('Saldo von "Old Anon User" vor dem Lauf', async ({ page }) => {
    await page.goto('/admin/users');

    const row = page.locator('#user-row-901');
    await expect(row).toBeVisible();
    await expect(row).toContainText('Old Anon User');

    const balanceCell = row.locator('td').nth(2);
    await expect(balanceCell).toHaveText(/5,00 EUR/);
  });

  test('Anonymisierungs-Modal zeigt fällige Zahlen und führt den Lauf aus', async ({ page }) => {
    await page.goto('/admin/settings');

    const anonymizeCard = page.locator('.card', { hasText: 'Anonymisierung (DSGVO)' });
    await anonymizeCard.locator('button', { hasText: 'Jetzt anonymisieren' }).click();

    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent).toContainText('4 Transaktionen von 1 Nutzern');

    const confirmBtn = modalContent.locator('button', { hasText: 'Anonymisieren (4 Transaktionen)' });
    await expect(confirmBtn).toBeVisible();
    await confirmBtn.click();

    // Modal is cleared client-side on a successful response.
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    const toastZone = page.locator('#toast-zone');
    await expect(toastZone.locator('.alert.alert-success')).toContainText(
      '4 Transaktionen von 1 Nutzern anonymisiert',
    );
  });

  test('Admin-Transaktionsliste zeigt anonymisierte Zeilen, frische Zeile bleibt zugeordnet', async ({ page }) => {
    // The four anonymized rows are 37 months old — the oldest rows in the
    // whole table by a wide margin — so under `ORDER BY created_at DESC`
    // they (plus the fresh -2,00 EUR row for the same user, seeded before
    // any test ran) sit at the very tail of the paginated list. Jump past
    // the end; the handler clamps to the actual last page.
    await page.goto('/admin/transactions?page=999999');

    const activePageBtn = page.locator('.join .btn-active');
    let lastPage = 1;
    if (await activePageBtn.count() > 0) {
      lastPage = parseInt((await activePageBtn.textContent())?.trim() || '1', 10);
    }
    const pagesToCheck = lastPage > 1 ? [lastPage, lastPage - 1] : [lastPage];

    let anonymizedBadgeCount = 0;
    let sawFreshOldAnonUserRow = false;

    for (const p of pagesToCheck) {
      await page.goto(`/admin/transactions?page=${p}`);
      const rows = page.locator('table tbody tr');
      const rowCount = await rows.count();
      for (let i = 0; i < rowCount; i++) {
        const text = (await rows.nth(i).textContent()) ?? '';
        if (text.includes('Anonymisiert')) {
          anonymizedBadgeCount++;
        }
        if (text.includes('Old Anon User') && text.includes('-2,00 EUR')) {
          sawFreshOldAnonUserRow = true;
        }
      }
    }

    expect(anonymizedBadgeCount).toBe(4);
    expect(sawFreshOldAnonUserRow).toBe(true);
  });

  test('Saldo von "Old Anon User" bleibt nach dem Lauf unverändert', async ({ page }) => {
    await page.goto('/admin/users');

    const row = page.locator('#user-row-901');
    await expect(row).toBeVisible();
    const balanceCell = row.locator('td').nth(2);
    await expect(balanceCell).toHaveText(/5,00 EUR/);
  });

  test('Protokoll-Tabelle zeigt den ausgeführten Lauf', async ({ page }) => {
    await page.goto('/admin/settings');

    const anonymizeCard = page.locator('.card', { hasText: 'Anonymisierung (DSGVO)' });
    const logRow = anonymizeCard.locator('table tbody tr').first();
    await expect(logRow).toBeVisible();

    const cells = logRow.locator('td');
    await expect(cells.nth(2)).toHaveText('4'); // Transaktionen
    await expect(cells.nth(3)).toHaveText('1'); // Nutzer
    await expect(cells.nth(4)).toHaveText(ADMIN_EMAIL); // Ausgeführt von
  });

  test('Erneutes Öffnen des Modals zeigt keine fälligen Transaktionen mehr', async ({ page }) => {
    await page.goto('/admin/settings');

    const anonymizeCard = page.locator('.card', { hasText: 'Anonymisierung (DSGVO)' });
    await anonymizeCard.locator('button', { hasText: 'Jetzt anonymisieren' }).click();

    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent).toContainText('0 Transaktionen von 0 Nutzern');

    // Count-gated confirm button must not be rendered when nothing is due.
    await expect(modalContent.locator('button', { hasText: /^Anonymisieren \(/ })).toHaveCount(0);

    await modalContent.locator('button', { hasText: 'Abbrechen' }).click();
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });
  });
});

test.describe('Admin Benutzerlöschung', () => {
  test('aktiver Benutzer hat keinen Löschen-Button', async ({ page }) => {
    await page.goto('/admin/users');

    // testuser is active and must not offer deletion. (Note: "Old Anon
    // User" / 901 is seeded deactivated, so it *does* show a "Löschen"
    // button — but its deletion is still blocked by a nonzero balance, see
    // the guard-order comment in the blocked-deletion test below.)
    const row = page.locator('tr', { hasText: 'testuser' }).first();
    await expect(row).toBeVisible();
    await expect(row.locator('button', { hasText: 'Löschen' })).toHaveCount(0);
  });

  test('deaktivierter Benutzer ohne Buchungen kann gelöscht werden', async ({ page }) => {
    await page.goto('/admin/users');

    const row = page.locator('#user-row-902');
    await expect(row).toBeVisible();
    await expect(row).toContainText('Delete Me');

    await row.locator('button', { hasText: 'Löschen' }).click();

    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent).toContainText('Delete Me');
    await expect(modalContent).toContainText('endgültig löschen');

    const confirmBtn = modalContent.locator('button.btn-error', { hasText: 'Löschen' });
    await expect(confirmBtn).toBeVisible();
    await confirmBtn.click();

    await expect(page.locator('#user-row-902')).toHaveCount(0, { timeout: 5000 });
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    // Persisted, not just a client-side removal.
    await page.reload();
    await expect(page.locator('#user-row-902')).toHaveCount(0);
  });

  test('Löschung eines Benutzers mit verbleibenden Buchungen ist blockiert', async ({ page }) => {
    await page.goto('/admin/users');

    const row = page.locator('#user-row-903');
    await expect(row).toBeVisible();
    await expect(row).toContainText('Delete Blocked');

    await row.locator('button', { hasText: 'Löschen' }).click();

    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible({ timeout: 5000 });

    // Blocked state: a guard message and no confirm ("Löschen") button, only
    // "Schließen". User 903 has a nonzero balance (its one -1,00 EUR
    // "Blocker" transaction), so the balance guard fires before the
    // remaining-references guard — see ConfirmDeleteUserModal's guard order
    // in internal/handler/admin/delete_user.go.
    await expect(modalContent).toContainText(
      /Kontostand muss 0,00 € betragen\.|Es existieren noch zugeordnete Buchungen/,
    );
    await expect(modalContent.locator('button.btn-error')).toHaveCount(0);
    await expect(modalContent.locator('button', { hasText: 'Schließen' })).toBeVisible();

    await modalContent.locator('button', { hasText: 'Schließen' }).click();
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    // Row must still exist (deletion did not happen).
    await expect(page.locator('#user-row-903')).toBeVisible();
  });
});
