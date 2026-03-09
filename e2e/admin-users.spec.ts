import { test, expect } from '@playwright/test';

test.describe('Admin User Management', () => {

  test('Benutzerliste ansehen', async ({ page }) => {
    await page.goto('/admin/users');

    const userList = page.locator('#user-list');
    await expect(userList).toBeVisible();

    const headers = userList.locator('thead th');
    await expect(headers).toContainText([
      'Name',
      'Email',
      'Guthaben',
      'Status',
      'Aktiv',
      'Limit-Override',
      'Aktionen',
    ]);

    const tableBody = userList.locator('tbody');
    await expect(tableBody).toContainText('testuser');
    await expect(tableBody).toContainText('testadmin');
  });

  test('Einzahlung buchen', async ({ page }) => {
    await page.goto('/admin/users');

    const userRow = page.locator('tr', { hasText: 'testuser' });
    await expect(userRow).toBeVisible();

    // Read balance before deposit
    const balanceBefore = await userRow.locator('td').nth(2).textContent();

    // Click deposit button
    const depositBtn = userRow.locator('.btn.btn-primary.btn-xs', { hasText: /Einzahlung/i });
    await depositBtn.click();

    // Wait for modal content to appear
    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible();
    await expect(modalContent).toContainText('Einzahlung für');

    // Fill deposit form
    const amountInput = page.locator('#deposit-amount');
    await amountInput.fill('10.00');
    const noteInput = page.locator('#deposit-note');
    await noteInput.fill('Test deposit');

    // Confirm deposit
    const confirmBtn = page.locator('#deposit-confirm-btn');
    await confirmBtn.click();

    // Wait for modal to close (content removed)
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    // Verify balance changed
    const updatedRow = page.locator('tr', { hasText: 'testuser' });
    const balanceAfter = await updatedRow.locator('td').nth(2).textContent();
    expect(balanceAfter).not.toBe(balanceBefore);
  });

  test('Barteamer-Status umschalten', async ({ page }) => {
    await page.goto('/admin/users');

    const userRow = page.locator('tr', { hasText: 'testuser' }).first();
    await expect(userRow).toBeVisible();

    // Get initial toggle state
    const barteamerToggle = userRow.locator('.toggle.toggle-primary');
    const wasChecked = await barteamerToggle.isChecked();

    // Click the barteamer toggle to open confirmation modal
    await barteamerToggle.click();

    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible({ timeout: 5000 });

    // Confirm the toggle
    const confirmBtn = modalContent.locator('.btn.btn-primary');
    await confirmBtn.click();

    // Wait for modal to close
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    // Verify toggle state changed
    const updatedRow = page.locator('tr', { hasText: 'testuser' }).first();
    const updatedToggle = updatedRow.locator('.toggle.toggle-primary');
    if (wasChecked) {
      await expect(updatedToggle).not.toBeChecked();
    } else {
      await expect(updatedToggle).toBeChecked();
    }

    // Toggle back to restore original state
    await updatedToggle.click();
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    const confirmBtn2 = modalContent.locator('.btn.btn-primary');
    await confirmBtn2.click();
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    const restoredRow = page.locator('tr', { hasText: 'testuser' }).first();
    const restoredToggle = restoredRow.locator('.toggle.toggle-primary');
    if (wasChecked) {
      await expect(restoredToggle).toBeChecked();
    } else {
      await expect(restoredToggle).not.toBeChecked();
    }
  });

  test('Benutzer deaktivieren/aktivieren', async ({ page }) => {
    await page.goto('/admin/users');

    const userRow = page.locator('tr', { hasText: 'testuser' }).first();
    await expect(userRow).toBeVisible();

    // Get initial active toggle state
    const activeToggle = userRow.locator('.toggle.toggle-success');
    const wasActive = await activeToggle.isChecked();

    // Deactivate user
    await activeToggle.click();

    const modalContent = page.locator('#modal .modal.modal-open');
    await expect(modalContent).toBeVisible({ timeout: 5000 });

    const confirmBtn = modalContent.locator('.btn.btn-primary');
    await confirmBtn.click();

    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    // Verify toggle state changed
    const updatedRow = page.locator('tr', { hasText: 'testuser' }).first();
    const updatedToggle = updatedRow.locator('.toggle.toggle-success');
    if (wasActive) {
      await expect(updatedToggle).not.toBeChecked();
    } else {
      await expect(updatedToggle).toBeChecked();
    }

    // Re-activate user to restore state
    await updatedToggle.click();
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    const confirmBtn2 = modalContent.locator('.btn.btn-primary');
    await confirmBtn2.click();
    await expect(modalContent).not.toBeVisible({ timeout: 5000 });

    const restoredRow = page.locator('tr', { hasText: 'testuser' }).first();
    const restoredToggle = restoredRow.locator('.toggle.toggle-success');
    if (wasActive) {
      await expect(restoredToggle).toBeChecked();
    } else {
      await expect(restoredToggle).not.toBeChecked();
    }

    // Verify admin cannot deactivate self
    const adminRow = page.locator('tr', { hasText: 'testadmin' });
    const adminActiveToggle = adminRow.locator('.toggle.toggle-success');
    await expect(adminActiveToggle).toBeDisabled();
  });
});
