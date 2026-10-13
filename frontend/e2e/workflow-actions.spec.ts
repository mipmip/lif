import { test, expect } from './fixtures';

test.describe('Workflow action buttons', () => {
  test('draft item shows Todo and Scrap buttons', async ({ lif, backlogPage, page }) => {
    lif.create('Draft Item', { status: 'draft', type: 'task' });

    await backlogPage.goto(1);
    await backlogPage.selectItem('Draft Item');

    const detail = page.locator('h1', { hasText: 'Draft Item' }).locator('..');

    await expect(detail.getByRole('button', { name: 'Todo' })).toBeVisible();
    await expect(detail.getByRole('button', { name: 'Scrap' })).toBeVisible();
    await expect(detail.getByRole('button', { name: 'Complete' })).not.toBeVisible();
  });

  test('todo item shows Scrap button (no Start Work without agent)', async ({
    lif,
    backlogPage,
    page
  }) => {
    lif.create('Todo Item', { status: 'todo', type: 'task' });

    await backlogPage.goto(1);
    await backlogPage.selectItem('Todo Item');

    const detail = page.locator('h1', { hasText: 'Todo Item' }).locator('..');

    await expect(detail.getByRole('button', { name: 'Scrap' })).toBeVisible();
    await expect(detail.getByRole('button', { name: 'Todo' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Complete' })).not.toBeVisible();
  });

  test('in-progress item shows Complete and Scrap buttons', async ({
    lif,
    boardPage,
    page
  }) => {
    lif.create('Active Item', { status: 'in-progress', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForItemInColumn('Active Item', 'in-progress');
    // Click the item in the board to select it
    await page.locator('[data-status="in-progress"] [role="listitem"]', { hasText: 'Active Item' }).locator('[role="button"]').click();

    const detail = page.locator('h1', { hasText: 'Active Item' }).locator('..');

    await expect(detail.getByRole('button', { name: 'Complete' })).toBeVisible();
    await expect(detail.getByRole('button', { name: 'Scrap' })).toBeVisible();
    await expect(detail.getByRole('button', { name: 'Todo' })).not.toBeVisible();
  });

  test('completed item shows no workflow buttons', async ({ lif, boardPage, page }) => {
    lif.create('Done Item', { status: 'completed', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForItemInColumn('Done Item', 'completed');
    await page.locator('[data-status="completed"] [role="listitem"]', { hasText: 'Done Item' }).locator('[role="button"]').click();

    const detail = page.locator('h1', { hasText: 'Done Item' }).locator('..');

    await expect(detail.getByRole('button', { name: 'Todo' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Scrap' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Complete' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Start Work' })).not.toBeVisible();
  });

  test('scrapped item shows no workflow buttons', async ({ lif, page }) => {
    const id = lif.create('Scrapped Item', { status: 'scrapped', type: 'task' });

    // Navigate directly with item param since scrapped items don't appear in any view
    await page.goto(`${lif.baseURL}/?item=${id}`);
    await expect(page.locator('h1', { hasText: 'Scrapped Item' })).toBeVisible({ timeout: 10_000 });

    const detail = page.locator('h1', { hasText: 'Scrapped Item' }).locator('..');

    await expect(detail.getByRole('button', { name: 'Todo' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Scrap' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Complete' })).not.toBeVisible();
  });

  test('Todo button moves draft item to todo status', async ({ lif, backlogPage, page }) => {
    lif.create('My Draft', { status: 'draft', type: 'task' });

    await backlogPage.goto(1);
    await backlogPage.selectItem('My Draft');

    const detail = page.locator('h1', { hasText: 'My Draft' }).locator('..');

    await detail.getByRole('button', { name: 'Todo' }).click();

    // Workflow buttons should update to todo state
    await expect(detail.getByRole('button', { name: 'Scrap' })).toBeVisible({ timeout: 5000 });
    await expect(detail.getByRole('button', { name: 'Todo' })).not.toBeVisible();
  });

  test('Complete button moves in-progress item to completed', async ({
    lif,
    boardPage,
    page
  }) => {
    lif.create('Active Task', { status: 'in-progress', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForItemInColumn('Active Task', 'in-progress');
    await page.locator('[data-status="in-progress"] [role="listitem"]', { hasText: 'Active Task' }).locator('[role="button"]').click();

    const detail = page.locator('h1', { hasText: 'Active Task' }).locator('..');

    await detail.getByRole('button', { name: 'Complete' }).click();

    // No workflow buttons should remain, Archive should appear
    await expect(detail.getByRole('button', { name: 'Archive' })).toBeVisible({ timeout: 5000 });
    await expect(detail.getByRole('button', { name: 'Complete' })).not.toBeVisible();
    await expect(detail.getByRole('button', { name: 'Scrap' })).not.toBeVisible();
  });

  test('Scrap button moves item to scrapped', async ({ lif, backlogPage, page }) => {
    lif.create('Unwanted Item', { status: 'todo', type: 'task' });

    await backlogPage.goto(1);
    await backlogPage.selectItem('Unwanted Item');

    const detail = page.locator('h1', { hasText: 'Unwanted Item' }).locator('..');

    await detail.getByRole('button', { name: 'Scrap' }).click();

    // No workflow buttons should remain, Archive should appear
    await expect(detail.getByRole('button', { name: 'Archive' })).toBeVisible({ timeout: 5000 });
    await expect(detail.getByRole('button', { name: 'Scrap' })).not.toBeVisible();
  });
});
