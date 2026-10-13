import { test, expect } from './fixtures';

test.describe('Board sorting', () => {
  test('items appear in the correct column by status', async ({ lif, boardPage }) => {
    lif.create('Todo Item', { status: 'todo', type: 'task' });
    lif.create('Active Item', { status: 'in-progress', type: 'task' });
    lif.create('Done Item', { status: 'completed', type: 'task' });

    await boardPage.goto();

    // Wait for all items to appear in their columns
    await boardPage.waitForItemInColumn('Todo Item', 'todo');
    await boardPage.waitForItemInColumn('Active Item', 'in-progress');
    await boardPage.waitForItemInColumn('Done Item', 'completed');

    expect(await boardPage.getColumnTitles('todo')).toEqual(['Todo Item']);
    expect(await boardPage.getColumnTitles('in-progress')).toEqual(['Active Item']);
    expect(await boardPage.getColumnTitles('completed')).toEqual(['Done Item']);
  });

  test('items within a column are sorted by priority', async ({ lif, boardPage }) => {
    lif.create('Low Task', { status: 'todo', priority: 'low', type: 'task' });
    lif.create('Critical Task', { status: 'todo', priority: 'critical', type: 'task' });
    lif.create('Normal Task', { status: 'todo', priority: 'normal', type: 'task' });
    lif.create('High Task', { status: 'todo', priority: 'high', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForColumnCount('todo', 4);

    const titles = await boardPage.getColumnTitles('todo');
    expect(titles).toEqual(['Critical Task', 'High Task', 'Normal Task', 'Low Task']);
  });

  test('item moves to new column when status changes on disk', async ({ lif, boardPage }) => {
    const id = lif.create('Moving Item', { status: 'todo', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForItemInColumn('Moving Item', 'todo');

    // Change status via CLI
    lif.update(id, { status: 'in-progress' });

    // Wait for it to appear in the new column
    await boardPage.waitForItemInColumn('Moving Item', 'in-progress');
    await boardPage.waitForItemNotInColumn('Moving Item', 'todo');
  });

  test('column re-sorts when item priority changes on disk', async ({ lif, boardPage }) => {
    const id = lif.create('Will Be Critical', {
      status: 'todo',
      priority: 'low',
      type: 'task'
    });
    lif.create('Normal Priority', { status: 'todo', priority: 'normal', type: 'task' });

    await boardPage.goto();
    await boardPage.waitForColumnCount('todo', 2);

    // Initially: Normal sorts before Low
    let titles = await boardPage.getColumnTitles('todo');
    expect(titles).toEqual(['Normal Priority', 'Will Be Critical']);

    // Promote to critical
    lif.update(id, { priority: 'critical' });

    // Should re-sort: critical before normal
    await expect(async () => {
      titles = await boardPage.getColumnTitles('todo');
      expect(titles).toEqual(['Will Be Critical', 'Normal Priority']);
    }).toPass({ timeout: 5_000 });
  });
});
