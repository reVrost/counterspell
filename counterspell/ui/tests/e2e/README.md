# E2E Tests

## Setup

Tests are powered by [Playwright](https://playwright.dev/).

Run once to install browsers:
```bash
cd ui && npx playwright install chromium
```

## Running Tests

### Option 1: Let Playwright start dev server
```bash
cd ui
npm run test:e2e
```

Playwright will automatically start the dev server on `:5173` and run tests.

### Option 2: Manual dev server (faster for repeated runs)
```bash
# Terminal 1: Start backend
make dev

# Terminal 2: Start frontend
cd ui && npm run dev

# Terminal 3: Run tests (reuse existing server)
cd ui && REUSE_SERVER=1 npm run test:e2e
```

### Run with UI (interactive)
```bash
cd ui && npm run test:e2e:ui
```

### View HTML report
```bash
cd ui && npm run test:e2e:report
```

## Test Files

- `dashboard.spec.ts` - Dashboard page tests

## Writing New Tests

### Text Assertions (Blind Agent Friendly)
```typescript
test('element has correct text', async ({ page }) => {
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');

  const title = page.locator('h1');
  await expect(title).toHaveText('Dashboard');
});
```

### Check for Console Errors
```typescript
test('no console errors', async ({ page }) => {
  const consoleErrors: string[] = [];

  page.on('console', msg => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text());
    }
  });

  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');

  expect(consoleErrors).toEqual([]);
});
```

### Check for Uncaught Exceptions
```typescript
test('no uncaught errors', async ({ page }) => {
  const uncaughtErrors: any[] = [];

  page.on('pageerror', error => {
    uncaughtErrors.push({
      message: error.message,
      name: error.name,
    });
  });

  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');

  expect(uncaughtErrors).toEqual([]);
});
```

### Check Element Visibility
```typescript
test('feed is visible', async ({ page }) => {
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');

  const feed = page.locator('#feed-content');
  await expect(feed).toBeVisible();
});
```

## Best Practices

1. **Use text assertions** - Works for blind agents
2. **Check console errors** - Catch runtime issues
3. **Use data-testid attributes** - Better selectors
4. **Wait for networkidle** - Ensure page fully loaded
5. **One assertion per test** - Easier to debug

## Configuration

`ui/playwright.config.ts`:
- Base URL: `http://localhost:5173`
- Browser: Chromium
- Reporter: List + HTML
- Retry: 0 locally, 2 in CI
