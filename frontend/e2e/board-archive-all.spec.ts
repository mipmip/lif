import { test, expect } from './fixtures';

test.describe('Board archive all', () => {
  test('archive all button is hidden when no completed items', async ({ lif, boardPage }) => {
    lif.create('Todo Item', { status: 'todo', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForColumnCount('todo', 1);

    await expect(boardPage.archiveAllButton).not.toBeVisible();
  });

  test('archive all button is visible when completed items exist', async ({ lif, boardPage }) => {
    lif.create('Done Item', { status: 'completed', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForColumnCount('completed', 1);

    await expect(boardPage.archiveAllButton).toBeVisible();
  });

  test('archive all removes all completed items from the board', async ({ lif, boardPage }) => {
    lif.create('Done A', { status: 'completed', type: 'task' });
    lif.create('Done B', { status: 'completed', type: 'task' });
    lif.create('Still Todo', { status: 'todo', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForColumnCount('completed', 2);
    await boardPage.waitForColumnCount('todo', 1);

    await boardPage.archiveAllCompleted();

    // Completed column should empty out
    await boardPage.waitForColumnCount('completed', 0);
    // Todo column should be unaffected
    await boardPage.waitForColumnCount('todo', 1);
  });

  test('cancel in confirmation modal does not archive items', async ({ lif, boardPage, page }) => {
    lif.create('Done Item', { status: 'completed', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForColumnCount('completed', 1);

    // Open the modal
    await boardPage.archiveAllButton.click();

    // Cancel
    await page.getByRole('button', { name: 'Cancel' }).click();

    // Item should still be there
    await boardPage.waitForColumnCount('completed', 1);
  });
});
