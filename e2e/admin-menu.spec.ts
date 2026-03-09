import { test, expect } from '@playwright/test';

test.describe('Admin Menu Management', () => {
  test('Kategorie erstellen — admin creates a new category', async ({ page }) => {
    await page.goto('/admin/menu');

    const categoryList = page.locator('#admin-category-list');
    await expect(categoryList).toBeVisible();

    // Fill in the new category form
    const createForm = page.locator('form[hx-post="/admin/categories"]');
    await createForm.locator('input[name="name"]').fill('Testkategorie');
    await createForm.locator('button[type="submit"]').click();

    // Expect the new category to appear in the list
    await expect(categoryList.getByText('Testkategorie')).toBeVisible();
  });

  test('Kategorie umbenennen — admin renames a category', async ({ page }) => {
    await page.goto('/admin/menu');

    const categoryList = page.locator('#admin-category-list');

    // Find the Testkategorie card
    const categoryCard = categoryList.locator('.card.bg-base-200', { hasText: 'Testkategorie' });
    await expect(categoryCard).toBeVisible();

    // Click the edit button for this category
    const editBtn = categoryCard.locator('[hx-get*="/admin/categories/"][hx-get*="/edit"]');
    await editBtn.click();

    // Wait for the modal with the edit input
    const nameInput = page.locator('#edit-category-name');
    await expect(nameInput).toBeVisible();

    // Clear and fill with new name
    await nameInput.clear();
    await nameInput.fill('Umbenannt');

    // Click confirm button
    const confirmBtn = page.locator('[hx-post*="/admin/categories/"][hx-post*="/update"]');
    await confirmBtn.click();

    // Expect the renamed category in the list
    await expect(categoryList.getByText('Umbenannt')).toBeVisible();
  });

  test('Getränk hinzufügen — admin adds an item to a category', async ({ page }) => {
    await page.goto('/admin/menu');

    const categoryList = page.locator('#admin-category-list');

    // Find the Umbenannt category card
    const categoryCard = categoryList.locator('.card.bg-base-200', { hasText: 'Umbenannt' });
    await expect(categoryCard).toBeVisible();

    // Find the add-item form within this category
    const addForm = categoryCard.locator('form[hx-post*="/admin/categories/"][hx-post*="/items"]');
    await addForm.locator('input[name="name"]').fill('Testgetränk');
    await addForm.locator('input[name="price_barteamer"]').fill('1.00');
    await addForm.locator('input[name="price_helfer"]').fill('1.50');
    await addForm.locator('button[type="submit"]').click();

    // Expect the new item to appear in the category
    await expect(categoryCard.getByText('Testgetränk')).toBeVisible();
  });

  test('Getränk bearbeiten — admin edits an item', async ({ page }) => {
    await page.goto('/admin/menu');

    const categoryList = page.locator('#admin-category-list');

    // Find the item in the list
    const categoryCard = categoryList.locator('.card.bg-base-200', { hasText: 'Umbenannt' });
    const itemRow = categoryCard.locator('ul.list').getByText('Testgetränk');
    await expect(itemRow).toBeVisible();

    // Click the edit button for this item
    const itemEntry = categoryCard.locator('ul.list li', { hasText: 'Testgetränk' });
    const editBtn = itemEntry.locator('[hx-get*="/admin/items/"][hx-get*="/edit"]');
    await editBtn.click();

    // Wait for the modal with edit inputs
    const nameInput = page.locator('#edit-item-name');
    await expect(nameInput).toBeVisible();
    await expect(page.locator('#edit-item-price-barteamer')).toBeVisible();
    await expect(page.locator('#edit-item-price-helfer')).toBeVisible();

    // Change the item name
    await nameInput.clear();
    await nameInput.fill('Bearbeitetes Getränk');

    // Click confirm
    const confirmBtn = page.locator('[hx-post*="/admin/items/"][hx-post*="/update"]');
    await confirmBtn.click();

    // Expect the updated item name in the list
    await expect(categoryCard.getByText('Bearbeitetes Getränk')).toBeVisible();
  });

  test('Getränk entfernen — admin soft-deletes an item', async ({ page }) => {
    await page.goto('/admin/menu');

    const categoryList = page.locator('#admin-category-list');
    const categoryCard = categoryList.locator('.card.bg-base-200', { hasText: 'Umbenannt' });

    // Find the item to delete
    const itemEntry = categoryCard.locator('ul.list li', { hasText: 'Bearbeitetes Getränk' });
    await expect(itemEntry).toBeVisible();

    // Override confirm to auto-accept
    await page.evaluate(() => { window.confirm = () => true; });

    // Click the delete button
    const deleteBtn = itemEntry.locator('[hx-post*="/admin/items/"][hx-post*="/delete"]');
    await deleteBtn.click();

    // Item should disappear from the list
    await expect(categoryCard.getByText('Bearbeitetes Getränk')).not.toBeVisible();
  });

  test('Kategorie löschen — admin deletes an empty category', async ({ page }) => {
    await page.goto('/admin/menu');

    const categoryList = page.locator('#admin-category-list');

    // Create a fresh empty category specifically for deletion testing
    // (the "Umbenannt" category has soft-deleted items with FK references)
    const createForm = page.locator('form[hx-post="/admin/categories"]');
    await createForm.locator('input[name="name"]').fill('Löschbar');
    await createForm.locator('button[type="submit"]').click();
    await expect(categoryList.getByText('Löschbar')).toBeVisible();

    // Now delete the empty category
    const categoryCard = categoryList.locator('.card.bg-base-200', { hasText: 'Löschbar' });
    await expect(categoryCard).toBeVisible();

    await page.evaluate(() => { window.confirm = () => true; });
    const deleteBtn = categoryCard.getByRole('button', { name: 'Entfernen' });
    await deleteBtn.click();

    // Category should disappear
    await expect(categoryList.getByText('Löschbar')).not.toBeVisible({ timeout: 10000 });
  });
});
