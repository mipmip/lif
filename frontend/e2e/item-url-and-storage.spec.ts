import type { Page } from '@playwright/test';
import { test, expect } from './fixtures';

/** Open the main workspace: the changes toggle lives in its ViewToolbar. */
async function openMainWorkspace(page: Page) {
  await page.locator('nav button:has(span.truncate)').first().click();
  await expect(page).toHaveURL(/\/workspace\//, { timeout: 10_000 });
}

/**
 * The selected item lives in the `item` query parameter and the layout/filter
 * state in `lif-*` localStorage keys. The old `bean` parameter and `beans-*`
 * keys are neither read nor migrated.
 */
test.describe('Item URL parameter', () => {
  test('?item=<id> preselects that item', async ({ lif, page }) => {
    const id = lif.create('Deep Linked', { status: 'todo', type: 'task' });

    await page.goto(`${lif.baseURL}/?item=${id}`);

    await expect(page.locator('h1', { hasText: 'Deep Linked' })).toBeVisible({ timeout: 10_000 });
  });

  test('?bean=<id> preselects nothing', async ({ lif, backlogPage, page }) => {
    const id = lif.create('Old Link', { status: 'todo', type: 'task' });

    await page.goto(`${lif.baseURL}/?bean=${id}`);
    await expect(backlogPage.itemRows).toHaveCount(1, { timeout: 10_000 });

    await expect(page.locator('h1', { hasText: 'Old Link' })).toHaveCount(0);
  });
});

test.describe('Persisted UI state keys', () => {
  test('the filter text is stored under lif-filter-text and restored from it', async ({
    lif,
    backlogPage,
    page
  }) => {
    lif.create('Alpha Item', { status: 'todo', type: 'task' });
    lif.create('Beta Item', { status: 'todo', type: 'task' });
    await backlogPage.goto(2);

    await page.getByTestId('filter-input').fill('alpha');
    await expect
      .poll(() => page.evaluate(() => localStorage.getItem('lif-filter-text')))
      .toBe('alpha');

    await page.reload();
    await expect(page.getByTestId('filter-input')).toHaveValue('alpha');
  });

  test('a stored beans-filter-text is ignored', async ({ lif, backlogPage, page }) => {
    lif.create('Alpha Item', { status: 'todo', type: 'task' });
    await backlogPage.goto(1);

    await page.evaluate(() => localStorage.setItem('beans-filter-text', 'nomatch'));
    await page.reload();

    await expect(backlogPage.itemRows).toHaveCount(1, { timeout: 10_000 });
    await expect(page.getByTestId('filter-input')).toHaveValue('');
  });

  test('the changes pane toggle is stored under lif-changes-pane and restored from it', async ({
    lif,
    backlogPage,
    page
  }) => {
    lif.create('Alpha Item', { status: 'todo', type: 'task' });
    await backlogPage.goto(1);
    await openMainWorkspace(page);

    await page.getByTitle('Show changes').click();
    await expect
      .poll(() => page.evaluate(() => localStorage.getItem('lif-changes-pane')))
      .toBe('true');

    await page.reload();
    await expect(page).toHaveURL(/\/workspace\//);
    await expect(page.getByTitle('Hide changes')).toBeVisible({ timeout: 10_000 });
  });

  test('a stored beans-changes-pane is ignored', async ({ lif, backlogPage, page }) => {
    lif.create('Alpha Item', { status: 'todo', type: 'task' });
    await backlogPage.goto(1);

    await openMainWorkspace(page);
    await page.evaluate(() => localStorage.setItem('beans-changes-pane', 'true'));
    await page.reload();

    await expect(page.getByTitle('Show changes')).toBeVisible({ timeout: 10_000 });
  });

  test('the sidebar width is restored from lif-split-sidebar', async ({
    lif,
    backlogPage,
    page
  }) => {
    lif.create('Alpha Item', { status: 'todo', type: 'task' });
    await backlogPage.goto(1);

    await page.evaluate(() => localStorage.setItem('lif-split-sidebar', '300'));
    await page.reload();

    await expect
      .poll(async () => (await page.locator('nav').first().boundingBox())?.width)
      .toBe(300);
  });

  test('a stored beans-split-sidebar is ignored', async ({ lif, backlogPage, page }) => {
    lif.create('Alpha Item', { status: 'todo', type: 'task' });
    await backlogPage.goto(1);

    await page.evaluate(() => localStorage.setItem('beans-split-sidebar', '300'));
    await page.reload();

    await expect
      .poll(async () => (await page.locator('nav').first().boundingBox())?.width)
      .toBe(224);
  });
});
