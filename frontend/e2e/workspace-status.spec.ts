import { test, expect } from './fixtures';

/**
 * Helper to create a workspace and return its name.
 */
async function createWorkspace(page: import('@playwright/test').Page) {
  const sidebar = page.locator('nav');
  await page.getByRole('button', { name: 'Create worktree' }).click();
  await expect(page).toHaveURL(/\/workspace\//, { timeout: 10_000 });

  const activeLabel = sidebar.locator('button.font-medium span.truncate');
  await expect(activeLabel).toBeVisible({ timeout: 5_000 });
  return (await activeLabel.textContent())!;
}

test.describe('Workspace status subscription', () => {
  test('shows "Ready to integrate" icon when worktree has uncommitted changes', async ({
    lif,
    page
  }) => {
    await page.goto(lif.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const wsName = await createWorkspace(page);

    // Create an uncommitted file in the worktree
    const worktrees = await lif.getWorktrees();
    const wt = worktrees.find((w) => w.branch.includes(wsName));
    expect(wt).toBeTruthy();
    lif.createFileInWorktree(wt!.path, 'dirty.txt', 'uncommitted content');

    // The subscription should deliver the status update within ~10s
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({ hasText: wsName });
    await expect(wsCard.locator('[title="Ready to integrate"]')).toBeVisible({ timeout: 15_000 });
  });

  test('shows "Ready to integrate" icon when worktree has unmerged commits', async ({
    lif,
    page
  }) => {
    await page.goto(lif.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const wsName = await createWorkspace(page);

    // Create a commit in the worktree
    const worktrees = await lif.getWorktrees();
    const wt = worktrees.find((w) => w.branch.includes(wsName));
    expect(wt).toBeTruthy();
    lif.commitInWorktree(wt!.path, 'committed.txt', 'committed content');

    // The subscription should deliver the status update within ~10s
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({ hasText: wsName });
    await expect(wsCard.locator('[title="Ready to integrate"]')).toBeVisible({ timeout: 15_000 });
  });

  test('does not show status icon for clean worktree', async ({ lif, page }) => {
    await page.goto(lif.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const wsName = await createWorkspace(page);

    // Wait long enough for at least one subscription cycle
    await page.waitForTimeout(2_000);

    // No status icon should be visible for a clean worktree
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({ hasText: wsName });
    await expect(wsCard.locator('[title="Ready to integrate"]')).not.toBeVisible();
    await expect(wsCard.locator('[title="Uncommitted changes"]')).not.toBeVisible();
  });
});
