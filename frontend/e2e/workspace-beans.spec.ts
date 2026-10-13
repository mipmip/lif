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
 * Helper: commit .lif/ directory to git so worktrees inherit bean files.
 */
function commitBeansDir(projectDir: string) {
  execFileSync('git', ['add', '.lif', '.lif.yml'], {
    cwd: projectDir,
    timeout: 10_000
  });
  execFileSync('git', ['commit', '-m', 'commit beans'], {
    cwd: projectDir,
    timeout: 10_000,
    env: GIT_ENV
  });
}

/**
 * Helper: find the bean file in a .lif/ directory by bean ID prefix.
 */
function findBeanFile(beansDir: string, beanId: string): string {
  const files = readdirSync(beansDir);
  const match = files.find((f) => f.startsWith(beanId) && f.endsWith('.md'));
  if (!match) throw new Error(`No bean file found for ${beanId} in ${beansDir}`);
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

test.describe('Workspace bean association', () => {
  test('beans modified in a worktree appear under the workspace in sidebar', async ({
    beans,
    page
  }) => {
    // Create a bean and commit .lif/ to git so worktrees inherit it
    const beanId = beans.create('WT Association Bean', { type: 'task', status: 'todo' });
    commitBeansDir(beans.projectDir);

    await page.goto(beans.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    // Create a workspace
    const { wsName, wtPath } = await createWorkspaceAndGetPath(page, () => beans.getWorktrees());

    // Modify the bean file in the worktree's .lif/ directory
    const beanFile = findBeanFile(join(wtPath, '.lif'), beanId);
    const beanPath = join(wtPath, '.lif', beanFile);
    const content = readFileSync(beanPath, 'utf-8');
    writeFileSync(beanPath, content.replace('status: todo', 'status: in-progress'));

    // The bean should appear under the workspace in the sidebar
    // (the file watcher detects the change, DetectBeanIDs picks it up via git diff)
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({
      has: page.locator('span.truncate', { hasText: wsName })
    });
    await expect(wsCard.getByText('WT Association Bean')).toBeVisible({ timeout: 10_000 });
  });

  test('committed bean changes in worktree appear under workspace', async ({ beans, page }) => {
    // Create a bean and commit .lif/ to git
    const beanId = beans.create('Committed Bean Change', { type: 'bug', status: 'todo' });
    commitBeansDir(beans.projectDir);

    await page.goto(beans.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const { wsName, wtPath } = await createWorkspaceAndGetPath(page, () => beans.getWorktrees());

    // Modify and commit the bean in the worktree
    const beanFile = findBeanFile(join(wtPath, '.lif'), beanId);
    const beanPath = join(wtPath, '.lif', beanFile);
    const content = readFileSync(beanPath, 'utf-8');
    writeFileSync(beanPath, content.replace('status: todo', 'status: in-progress'));
    execFileSync('git', ['add', '.lif'], { cwd: wtPath, timeout: 10_000 });
    execFileSync('git', ['commit', '-m', 'update bean status'], {
      cwd: wtPath,
      timeout: 10_000,
      env: GIT_ENV
    });

    // The bean should appear under the workspace
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({
      has: page.locator('span.truncate', { hasText: wsName })
    });
    await expect(wsCard.getByText('Committed Bean Change')).toBeVisible({ timeout: 10_000 });
  });

  test('unmodified beans do not appear under workspace', async ({ beans, page }) => {
    // Create a bean and commit .lif/ to git
    beans.create('Unchanged Bean', { type: 'task', status: 'todo' });
    commitBeansDir(beans.projectDir);

    await page.goto(beans.baseURL + '/');
    await expect(page.getByText('Workspaces')).toBeVisible({ timeout: 10_000 });

    const { wsName } = await createWorkspaceAndGetPath(page, () => beans.getWorktrees());

    // Don't modify the bean — it should NOT appear under the workspace
    const sidebar = page.locator('nav');
    const wsCard = sidebar.locator('div.rounded-md').filter({
      has: page.locator('span.truncate', { hasText: wsName })
    });

    // Wait a moment to ensure the subscription has settled, then verify absence
    await page.waitForTimeout(2_000);
    await expect(wsCard.getByText('Unchanged Bean')).not.toBeVisible();
  });
});
