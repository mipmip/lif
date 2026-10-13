import { describe, it, expect } from 'vitest';
import { renderMarkdown, linkifyItemIds } from './markdown';

// Item references were rendered as <a> without href: not focusable, not
// reachable by keyboard, and no way to open one in a new tab. The click
// handler stays, but it now enhances a real link instead of replacing it.
describe('item references', () => {
  it('render as focusable links in markdown', async () => {
    const html = await renderMarkdown('see beans-a1b2 for details');

    expect(html).toContain('href="?item=beans-a1b2"');
  });

  it('render as focusable links in plain text', () => {
    const html = linkifyItemIds('see beans-a1b2');

    expect(html).toContain('href="?item=beans-a1b2"');
  });
});

describe('item reference link hooks', () => {
  it('carry the item id in data-item-id', async () => {
    const html = await renderMarkdown('see beans-a1b2');

    expect(html).toContain('data-item-id="beans-a1b2"');
  });

  it('carry the item-link class', async () => {
    const html = await renderMarkdown('see beans-a1b2');

    expect(html).toContain('class="item-link"');
  });
});
