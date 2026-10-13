import { expect, type Locator, type Page } from '@playwright/test';

/**
 * Page object for the backlog (list) view at /.
 */
export class BacklogPage {
  readonly itemRows: Locator;

  constructor(
    private page: Page,
    private baseURL: string
  ) {
    this.itemRows = page.locator('.item-row');
  }

  /**
   * Navigate to the backlog page and wait for items to load.
   * @param expectedCount If provided, wait until exactly this many items are visible.
   */
  async goto(expectedCount?: number) {
    await this.page.goto(this.baseURL + '/');
    if (expectedCount !== undefined && expectedCount > 0) {
      await expect(this.itemRows).toHaveCount(expectedCount, { timeout: 10_000 });
    } else if (expectedCount === undefined) {
      await this.page.waitForSelector('.item-row', { timeout: 10_000 });
    }
  }

  /** Get all visible item titles in display order. */
  async getItemTitles(): Promise<string[]> {
    const titles = await this.itemRows
      .locator('[role="button"] > div > span.text-sm')
      .allTextContents();
    return titles.map((t) => t.trim());
  }

  /** Get all visible item statuses in display order. */
  async getItemStatuses(): Promise<string[]> {
    const statuses = await this.itemRows
      .locator('[role="button"] > div > span.rounded-full')
      .allTextContents();
    return statuses.map((s) => s.trim());
  }

  /** Click on an item by its title. */
  async selectItem(title: string) {
    await this.itemRows.filter({ hasText: title }).first().locator('[role="button"]').click();
  }

  /** Wait for a specific item to appear. */
  async waitForItem(title: string) {
    await this.page.locator('.item-row', { hasText: title }).waitFor({ timeout: 10_000 });
  }

  /** Wait for an item to disappear from the list. */
  async waitForItemGone(title: string) {
    await this.page
      .locator('.item-row', { hasText: title })
      .waitFor({ state: 'detached', timeout: 10_000 });
  }

  /** Get the count of visible items. */
  async count(): Promise<number> {
    return this.itemRows.count();
  }

  /** Get the .item-row for a specific item by title (uses data-item-id for precision). */
  itemByTitle(title: string): Locator {
    return this.itemRows.filter({ hasText: title }).first();
  }

  /**
   * Get the draggable card element for an item, identified by title.
   * Each [draggable] div contains only its own card's content (not descendants),
   * so filtering by text gives us the exact item's drag handle.
   */
  private draggableByTitle(title: string): Locator {
    return this.page.locator(
      `[draggable="true"]:has([role="button"] span.text-sm:text-is("${title}"))`
    );
  }

  /**
   * Drag an item to reorder it above/below another item, or onto it to reparent.
   *
   * The drop zones are: top 25% = above, middle 50% = reparent, bottom 25% = below.
   * We target 10%/90% for reorder and 50% for reparent to avoid zone boundaries.
   */
  async dragItem(
    sourceTitle: string,
    targetTitle: string,
    position: 'above' | 'below' | 'onto' = 'above'
  ) {
    const source = this.draggableByTitle(sourceTitle);
    const target = this.draggableByTitle(targetTitle);

    const targetBox = await target.boundingBox();
    if (!targetBox) throw new Error(`Target item "${targetTitle}" not visible`);

    // Compute Y offset within the target card
    const yFraction = position === 'above' ? 0.1 : position === 'below' ? 0.9 : 0.5;
    const targetY = targetBox.height * yFraction;

    await source.dragTo(target, {
      targetPosition: { x: targetBox.width / 2, y: targetY }
    });
  }
}
