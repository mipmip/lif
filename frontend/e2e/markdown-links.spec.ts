import { test, expect } from './fixtures';

test.describe('Markdown links', () => {
  test('external links have target="_blank" and rel="noopener noreferrer"', async ({
    lif,
    backlogPage,
    page
  }) => {
    const id = lif.create('Link Test Item', {
      status: 'todo',
      type: 'task'
    });

    // Update the item body with a markdown link and a bare URL
    lif.run([
      'update',
      id,
      '--body-append',
      'Visit [Example](https://example.com) or https://example.org for more info.'
    ]);

    await backlogPage.goto(1);
    await backlogPage.selectItem('Link Test Item');

    // Wait for the item detail body to render
    const body = page.locator('.item-body');
    await expect(body).toBeVisible({ timeout: 10_000 });

    // Check the explicit markdown link
    const markdownLink = body.locator('a[href="https://example.com"]');
    await expect(markdownLink).toHaveAttribute('target', '_blank');
    await expect(markdownLink).toHaveAttribute('rel', 'noopener noreferrer');

    // Check the autolinked bare URL
    const autoLink = body.locator('a[href="https://example.org"]');
    await expect(autoLink).toHaveAttribute('target', '_blank');
    await expect(autoLink).toHaveAttribute('rel', 'noopener noreferrer');
  });
});
