import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test('page loads without errors', async ({ page }) => {
    const consoleErrors: string[] = [];

    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    expect(consoleErrors).toEqual([]);

    const title = await page.title();
    expect(title).toContain('Dashboard');
  });

  test('feed section is visible', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const feedContent = page.locator('#feed-content');
    await expect(feedContent).toBeVisible();
  });

  test('loading state appears initially', async ({ page }) => {
    await page.goto('/dashboard');

    const loadingText = page.getByTestId('loading-state').getByText('Loading feed...');
    await expect(loadingText).toBeVisible();
  });

  test('console logs contain no uncaught errors', async ({ page }) => {
    const uncaughtErrors: any[] = [];

    page.on('pageerror', error => {
      uncaughtErrors.push({
        message: error.message,
        name: error.name,
      });
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    if (uncaughtErrors.length > 0) {
      console.error('Uncaught errors:', uncaughtErrors);
    }

    expect(uncaughtErrors).toEqual([]);
  });

  test('page has accessible heading', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const heading = page.locator('h1, h2').first();
    await expect(heading).toBeVisible();
  });

  test('feed loaded state appears', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    const feedLoaded = page.getByTestId('feed-loaded');
    const feedContent = page.locator('#feed-content');

    await expect(feedLoaded).toBeVisible();
    await expect(feedContent).toBeVisible();
  });
});
