import { test, expect } from './fixtures';

/**
 * Item references in rendered markdown are real anchors (`href="?item=…"`), so
 * keyboard users reach them. The itemLinks action intercepts the activation and
 * selects the item in place.
 *
 * A reference whose item is not in the store — archived, deleted, from another
 * store, or a false positive of the `beans-xxxx` matcher — must do nothing at
 * all. If the action returns without preventDefault, the browser follows the
 * anchor to `?item=<id>`; the selection does not follow (the layout applies
 * `selectedItemId` exactly once), so the URL ends up advertising an item that is
 * neither selected nor existent, and a reload lands on a dangling selection.
 */
test.describe('Item reference links', () => {
  test('a reference to an item that is not in the store does not navigate', async ({
    lif,
    backlogPage,
    page
  }) => {
    const hostId = lif.create('Host Item', { status: 'todo', type: 'task' });
    // Shaped like an item reference so the markdown matcher linkifies it, but no
    // such item exists in this store.
    const danglingId = 'beans-zzzz';
    lif.run(['update', hostId, '--body-append', `Superseded by ${danglingId}.`]);

    await backlogPage.goto(1);
    await backlogPage.selectItem('Host Item');

    const body = page.locator('.item-body');
    await expect(body).toBeVisible({ timeout: 10_000 });

    const link = body.locator(`a[data-item-id="${danglingId}"]`);
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', `?item=${danglingId}`);

    const urlBefore = page.url();
    const historyBefore = await page.evaluate(() => history.length);
    await link.click();

    // The click is a no-op: the browser must not follow the anchor. The URL
    // alone is not enough evidence — uiState rewrites `?item=` back to the
    // selected item via replaceState, which hides the navigation. The pushed
    // history entry does not get rewritten, so that is what pins it.
    await page.waitForTimeout(500);
    expect(await page.evaluate(() => history.length)).toBe(historyBefore);
    expect(page.url()).toBe(urlBefore);
  });
});
