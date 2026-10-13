import { test, expect } from './fixtures';

test.describe('Backlog sorting', () => {
  test('items are sorted by priority, then type, then title within each section', async ({
    lif,
    backlogPage
  }) => {
    lif.create('Todo Normal Task', { status: 'todo', priority: 'normal', type: 'task' });
    lif.create('Todo Normal Bug', { status: 'todo', priority: 'normal', type: 'bug' });
    lif.create('Todo High Feature', { status: 'todo', priority: 'high', type: 'feature' });
    lif.create('Draft Idea', { status: 'draft', priority: 'low', type: 'task' });

    await backlogPage.goto(4);

    const titles = await backlogPage.getItemTitles();
    // Todo section first (sorted by priority, then type, then title), then Draft section
    expect(titles).toEqual([
      'Todo High Feature',
      'Todo Normal Bug',
      'Todo Normal Task',
      'Draft Idea'
    ]);
  });

  test('list re-sorts when an item priority changes on disk', async ({ lif, backlogPage }) => {
    const id1 = lif.create('Low Priority Task', {
      status: 'todo',
      priority: 'low',
      type: 'task'
    });
    lif.create('Normal Priority Task', { status: 'todo', priority: 'normal', type: 'task' });

    await backlogPage.goto(2);

    let titles = await backlogPage.getItemTitles();
    expect(titles).toEqual(['Normal Priority Task', 'Low Priority Task']);

    // Change the low-priority item to critical via CLI (filesystem change)
    lif.update(id1, { priority: 'critical' });

    // The item should now appear first
    await expect(async () => {
      titles = await backlogPage.getItemTitles();
      expect(titles).toEqual(['Low Priority Task', 'Normal Priority Task']);
    }).toPass({ timeout: 5_000 });
  });

  test('list re-sorts when an item status changes on disk', async ({ lif, backlogPage }) => {
    const id1 = lif.create('A Draft Item', { status: 'draft', type: 'task' });
    lif.create('B Todo Item', { status: 'todo', type: 'task' });

    await backlogPage.goto(2);

    // Todo section comes before Draft section
    let titles = await backlogPage.getItemTitles();
    expect(titles).toEqual(['B Todo Item', 'A Draft Item']);

    // Move the draft item to todo
    lif.update(id1, { status: 'todo' });

    // Both are now todo, should sort by title
    await expect(async () => {
      titles = await backlogPage.getItemTitles();
      expect(titles).toEqual(['A Draft Item', 'B Todo Item']);
    }).toPass({ timeout: 5_000 });
  });

  test('new item appears in correct sorted position', async ({ lif, backlogPage }) => {
    lif.create('Zebra Task', { status: 'todo', type: 'task' });
    lif.create('Alpha Task', { status: 'todo', type: 'task' });

    await backlogPage.goto(2);

    let titles = await backlogPage.getItemTitles();
    expect(titles).toEqual(['Alpha Task', 'Zebra Task']);

    // Create a new item that should sort between the two
    lif.create('Middle Task', { status: 'todo', type: 'task' });

    await expect(async () => {
      titles = await backlogPage.getItemTitles();
      expect(titles).toEqual(['Alpha Task', 'Middle Task', 'Zebra Task']);
    }).toPass({ timeout: 5_000 });
  });

  test('dragging an item reorders it within the backlog', async ({ lif, backlogPage }) => {
    // Create items with same status/priority/type so they sort by title
    lif.create('Alpha', { status: 'todo', type: 'task' });
    lif.create('Bravo', { status: 'todo', type: 'task' });
    lif.create('Charlie', { status: 'todo', type: 'task' });

    await backlogPage.goto(3);

    // Initial order: Alpha, Bravo, Charlie
    let titles = await backlogPage.getItemTitles();
    expect(titles).toEqual(['Alpha', 'Bravo', 'Charlie']);

    // Drag Charlie above Alpha
    await backlogPage.dragItem('Charlie', 'Alpha', 'above');

    // New order: Charlie, Alpha, Bravo
    await expect(async () => {
      titles = await backlogPage.getItemTitles();
      expect(titles).toEqual(['Charlie', 'Alpha', 'Bravo']);
    }).toPass({ timeout: 5_000 });
  });

  test('dragging an item onto another reparents it and persists', async ({
    lif,
    backlogPage,
    page
  }) => {
    lif.create('Parent Item', { status: 'todo', type: 'feature' });
    lif.create('Child Item', { status: 'todo', type: 'task' });

    await backlogPage.goto(2);

    // Both are top-level initially
    let titles = await backlogPage.getItemTitles();
    expect(titles).toContain('Parent Item');
    expect(titles).toContain('Child Item');

    // Drag Child Item onto Parent Item to reparent it
    await backlogPage.dragItem('Child Item', 'Parent Item', 'onto');

    // Child Item should now be nested under Parent Item
    await expect(async () => {
      const parentItem = backlogPage.itemByTitle('Parent Item');
      const nestedChild = parentItem.locator('.item-row', { hasText: 'Child Item' });
      await expect(nestedChild).toBeVisible();
    }).toPass({ timeout: 5_000 });

    // Reload the page to verify persistence
    await page.reload();
    await backlogPage.waitForItem('Parent Item');

    // After reload, Child Item should still be nested under Parent Item
    await expect(async () => {
      const parentItem = backlogPage.itemByTitle('Parent Item');
      const nestedChild = parentItem.locator('.item-row', { hasText: 'Child Item' });
      await expect(nestedChild).toBeVisible();
    }).toPass({ timeout: 5_000 });
  });

  test('dragging an item into a specific position within another parent works', async ({
    lif,
    backlogPage,
    page
  }) => {
    // Create a feature with 3 child tasks
    const parentId = lif.create('Parent Feature', { status: 'todo', type: 'feature' });
    const childA = lif.create('Alpha Child', { status: 'todo', type: 'task' });
    const childB = lif.create('Bravo Child', { status: 'todo', type: 'task' });
    const childC = lif.create('Charlie Child', { status: 'todo', type: 'task' });

    // Set parent relationships via CLI
    lif.run(['update', childA, '--parent', parentId]);
    lif.run(['update', childB, '--parent', parentId]);
    lif.run(['update', childC, '--parent', parentId]);

    // Create a top-level task to drag in
    lif.create('Interloper', { status: 'todo', type: 'task' });

    await backlogPage.goto(5); // Parent Feature + 3 children + Interloper

    // Drag Interloper above Bravo Child (between Alpha and Bravo)
    await backlogPage.dragItem('Interloper', 'Bravo Child', 'above');

    // Verify: Interloper should be between Alpha Child and Bravo Child
    await expect(async () => {
      const parentItem = backlogPage.itemByTitle('Parent Feature');
      const childTitles = await parentItem
        .locator('.item-row [role="button"] > div > span.text-sm')
        .allTextContents();
      const trimmed = childTitles.map((t) => t.trim());
      expect(trimmed).toEqual(['Alpha Child', 'Interloper', 'Bravo Child', 'Charlie Child']);
    }).toPass({ timeout: 5_000 });

    // Verify persistence
    await page.reload();
    await backlogPage.waitForItem('Parent Feature');

    await expect(async () => {
      const parentItem = backlogPage.itemByTitle('Parent Feature');
      const childTitles = await parentItem
        .locator('.item-row [role="button"] > div > span.text-sm')
        .allTextContents();
      const trimmed = childTitles.map((t) => t.trim());
      expect(trimmed).toEqual(['Alpha Child', 'Interloper', 'Bravo Child', 'Charlie Child']);
    }).toPass({ timeout: 5_000 });
  });

  test('deleted item disappears from list', async ({ lif, backlogPage }) => {
    const id1 = lif.create('Item To Delete', { status: 'todo', type: 'task' });
    lif.create('Item To Keep', { status: 'todo', type: 'task' });

    await backlogPage.goto(2);

    // Delete the item via CLI
    lif.run(['delete', '--force', id1]);

    await backlogPage.waitForItemGone('Item To Delete');
    expect(await backlogPage.count()).toBe(1);
  });
});
