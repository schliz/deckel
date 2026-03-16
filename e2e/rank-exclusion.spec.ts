import { test, expect } from '@playwright/test';

// This test runs as a regular user (user-tests project) to verify
// that kiosk users are not counted in the leaderboard rank total.

test.describe('Kiosk Rank Exclusion', () => {

  test('Rangliste — kiosk user not counted in rank total', async ({ page }) => {
    await page.goto('/');

    // Wait for header stats to load (loaded lazily via hx-get)
    const headerStats = page.locator('#header-stats');
    await expect(headerStats).toContainText('Platz', { timeout: 10000 });

    // Extract the rank text "Platz X von Y"
    const rankText = await headerStats.textContent();
    const match = rankText?.match(/Platz\s+(\d+)\s+von\s+(\d+)/);
    expect(match).toBeTruthy();

    const total = parseInt(match![2]);

    // With testuser and testadmin as regular users, and testkiosk excluded,
    // the total should be 2 (not 3)
    expect(total).toBe(2);
  });

});
