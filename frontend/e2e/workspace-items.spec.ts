import { execFileSync } from 'node:child_process';
import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { test, expect } from './fixtures';

const GIT_ENV = {
  ...process.env,
  GIT_AUTHOR_NAME: 'test',
  GIT_AUTHOR_EMAIL: 'test@test',
  GIT_COMMITTER_NAME: 'test',
  GIT_COMMITTER_EMAIL: 'test@test'
};

/**
 * Helper: commit .lif/ directory to git so worktrees inherit item files.
 */
function commitItemsDir(projectDir: string) {
  execFileSync('git', ['add', '.lif', '.lif.yml'], {
    cwd: projectDir,
    timeout: 10_000
  });
  execFileSync('git', ['commit', '-m', 'commit items'], {
    cwd: projectDir,
    timeout: 10_000,
    env: GIT_ENV
  });
}

/**
 * Helper: find the item file in a .lif/ directory by item ID prefix.
 */
function findItemFile(itemsDir: string, itemId: string): string {
  const files = readdirSync(itemsDir);
  const match = files.find((f) => f.startsWith(itemId) && f.endsWith('.md'));
  if (!match) throw new Error(`No item file found for ${itemId} in ${itemsDir}`);
  return match;
}

/**
 * Helper: create a workspace via the UI and return its name + worktree path.
 */
async function createWorkspaceAndGetPath(
  page: import('@playwright/test').Page,
  getWorktrees: () => Promise<{ id: string; path: string; branch: string }[]>
) {
  const sidebar = page.locator('nav');
  await page.getByRole('button', { name: 'Create worktree' }).click();
  await expect(page).toHaveURL(/\/workspace\//, { timeout: 10_000 });

  const activeLabel = sidebar.locator('button.font-medium span.truncate');
  await expect(activeLabel).toBeVisible({ timeout: 5_000 });
  const wsName = (await activeLabel.textContent())!;

  const worktrees = await getWorktrees();
  const wt = worktrees.find((w) => w.branch.includes(wsName));
  expect(wt).toBeTruthy();

  return { wsName, wtPath: wt!.path };
}

test.describe('Workspace item association', () => {
  test('items modified in a worktree appear under the workspace in sidebar', async ({
    lif,
    page
  }) => {
    // Create an item and commit .lif/ to git so worktrees inherit it
    const itemId = lif.create('WT Association Item', { type: 'task', status: 'todo' });
    commitItemsDir(lif.projectDir);

    await page.goto(lif.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    // Create a workspace
    const { wsName, wtPath } = await createWorkspaceAndGetPath(page, () => lif.getWorktrees());

    // Modify the item file in the worktree's .lif/ directory
    const itemFile = findItemFile(join(wtPath, '.lif'), itemId);
    const itemPath = join(wtPath, '.lif', itemFile);
    const content = readFileSync(itemPath, 'utf-8');
    writeFileSync(itemPath, content.replace('status: todo', 'status: in-progress'));

    // The item should appear under the workspace in the sidebar
    // (the file watcher detects the change, DetectItemIDs picks it up via git diff)
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({
      has: page.locator('span.truncate', { hasText: wsName })
    });
    await expect(wsCard.getByText('WT Association Item')).toBeVisible({ timeout: 10_000 });
  });

  test('committed item changes in worktree appear under workspace', async ({ lif, page }) => {
    // Create an item and commit .lif/ to git
    const itemId = lif.create('Committed Item Change', { type: 'bug', status: 'todo' });
    commitItemsDir(lif.projectDir);

    await page.goto(lif.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const { wsName, wtPath } = await createWorkspaceAndGetPath(page, () => lif.getWorktrees());

    // Modify and commit the item in the worktree
    const itemFile = findItemFile(join(wtPath, '.lif'), itemId);
    const itemPath = join(wtPath, '.lif', itemFile);
    const content = readFileSync(itemPath, 'utf-8');
    writeFileSync(itemPath, content.replace('status: todo', 'status: in-progress'));
    execFileSync('git', ['add', '.lif'], { cwd: wtPath, timeout: 10_000 });
    execFileSync('git', ['commit', '-m', 'update item status'], {
      cwd: wtPath,
      timeout: 10_000,
      env: GIT_ENV
    });

    // The item should appear under the workspace
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({
      has: page.locator('span.truncate', { hasText: wsName })
    });
    await expect(wsCard.getByText('Committed Item Change')).toBeVisible({ timeout: 10_000 });
  });

  test('unmodified items do not appear under workspace', async ({ lif, page }) => {
    // Create an item and commit .lif/ to git
    lif.create('Unchanged Item', { type: 'task', status: 'todo' });
    commitItemsDir(lif.projectDir);

    await page.goto(lif.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const { wsName } = await createWorkspaceAndGetPath(page, () => lif.getWorktrees());

    // Don't modify the item — it should NOT appear under the workspace
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({
      has: page.locator('span.truncate', { hasText: wsName })
    });

    // Wait a moment to ensure the subscription has settled, then verify absence
    await page.waitForTimeout(2_000);
    await expect(wsCard.getByText('Unchanged Item')).not.toBeVisible();
  });
});
