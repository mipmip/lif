import { test, expect } from './fixtures';

/**
 * The per-card controls inside a backlog item card (Archive, "go to workspace")
 * live in ItemCard, which is rendered inside ItemRow's drag wrapper. The
 * wrapper also handles clicks, to select the item. Both are delegated handlers,
 * so the inner button wins and its stopPropagation keeps the wrapper out of it.
 *
 * Wiring the wrapper's click through a `use:` action breaks that: a direct
 * listener fires during real bubbling, before Svelte's delegation walk ever
 * starts, so the wrapper's stopPropagation silently disables every button in
 * the card. This test pins the button, not the wiring.
 */
test.describe('Backlog card actions', () => {
  test('the archive button on a backlog card archives the item', async ({
    lif,
    backlogPage,
    page
  }) => {
    // Only todo/draft items are listed at top level, but a card's children are
    // listed whatever their status — so a completed child is where the Archive
    // button is actually reachable in the backlog.
    const parentId = lif.create('Parent Item', { status: 'todo', type: 'epic' });
    const childId = lif.create('Done Child', { status: 'todo', type: 'task' });
    lif.run(['update', childId, '--parent', parentId]);
    lif.update(childId, { status: 'completed' });

    await backlogPage.goto(2);

    const child = page.locator(`.item-row[data-item-id="${childId}"]`);
    await expect(child).toBeVisible();

    await child.getByRole('button', { name: 'Archive', exact: true }).click();

    // The item must actually leave the list. Before the fix the click merely
    // selected the item and the card stayed put.
    await expect(child).toHaveCount(0, { timeout: 10_000 });
    await expect(backlogPage.itemRows).toHaveCount(1);
  });
});
