import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { test, expect } from './fixtures';
import { agentSession } from './agent-session';

test.describe('Agent chat', () => {
  test('Item IDs are autolinked in tool and info messages', async ({ page, lif }) => {
    await agentSession('__central__', lif)
      .withMessages([
        { role: 'user', content: 'update the item' },
        { role: 'tool', content: 'Bash: lif update beans-ab12 -s in-progress' },
        { role: 'info', content: 'Working on beans-cd34' },
        { role: 'assistant', content: 'Done updating beans-ab12.' }
      ])
      .open(page);

    // Tool message should have an autolinked item ID
    const toolItemLink = page.locator('.item-link[data-item-id="beans-ab12"]').first();
    await expect(toolItemLink).toBeVisible({ timeout: 5000 });

    // Info message should have an autolinked item ID
    const infoItemLink = page.locator('.item-link[data-item-id="beans-cd34"]');
    await expect(infoItemLink).toBeVisible({ timeout: 5000 });

    // Assistant message should also have an autolinked item ID (via renderMarkdown)
    const assistantItemLink = page.locator('.item-link[data-item-id="beans-ab12"]').last();
    await expect(assistantItemLink).toBeVisible({ timeout: 5000 });
  });

  test('Clear button resets the conversation in the UI', async ({ page, lif }) => {
    await agentSession('__central__', lif)
      .withMessages([
        { role: 'user', content: 'hello agent' },
        { role: 'assistant', content: 'Hi! How can I help?' }
      ])
      .open(page);

    // Verify the seeded messages are visible
    await expect(page.locator('text=hello agent')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('text=Hi! How can I help?')).toBeVisible({ timeout: 5000 });

    // Clear button should be enabled
    const clearBtn = page.locator('button:has-text("Clear")');
    await expect(clearBtn).toBeEnabled();

    // Click Clear
    await clearBtn.click();

    // The empty state message should reappear
    await expect(page.locator('text=Send a message to start a conversation')).toBeVisible({
      timeout: 5000
    });

    // The messages should be gone
    await expect(page.locator('text=hello agent')).not.toBeVisible();
    await expect(page.locator('text=Hi! How can I help?')).not.toBeVisible();

    // Clear button should be disabled again
    await expect(clearBtn).toBeDisabled();
  });

  test('Image attachments are displayed inline in user messages', async ({ page, lif }) => {
    // Create a tiny 1x1 red PNG for testing
    const pngData = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==',
      'base64'
    );
    const imageId = 'test-image.png';

    // Write the image file to the attachments directory
    const attachDir = join(lif.lifPath, '.conversations', 'attachments', '__central__');
    mkdirSync(attachDir, { recursive: true });
    writeFileSync(join(attachDir, imageId), pngData);

    // Seed a conversation with an image reference
    await agentSession('__central__', lif)
      .withMessages([
        {
          role: 'user',
          content: 'Check this screenshot',
          images: [{ id: imageId, media_type: 'image/png' }]
        },
        { role: 'assistant', content: 'I can see the image.' }
      ])
      .open(page);

    // Verify the text message is visible
    await expect(page.locator('text=Check this screenshot')).toBeVisible({ timeout: 5000 });

    // Verify the image is rendered inline
    const img = page.locator('img[alt="Attachment"]');
    await expect(img).toBeVisible({ timeout: 5000 });

    // Verify the image src points to the attachment endpoint
    const src = await img.getAttribute('src');
    expect(src).toContain('/api/attachments/__central__/test-image.png');
  });
});
