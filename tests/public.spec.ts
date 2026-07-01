import { expect, test } from '@playwright/test';

import { createPost, login, uniqueId } from './helpers';

test.describe.serial('public post page', () => {
  test('shows post detail with title, content, and tags', async ({ page }) => {
    await login(page);
    const title = uniqueId('detail-post');
    const tag = uniqueId('detail-tag');

    await createPost(page, {
      title,
      content: 'Detail page body',
      tags: tag,
    });

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await page.locator('.post').filter({ hasText: title }).getByRole('link').first().click();

    await expect(page).toHaveURL(/\/posts\/\d+/);
    await expect(page.locator('.post-title')).toContainText(title);
    await expect(page.locator('.post-content')).toContainText('Detail page body');
    await expect(page.getByRole('link', { name: tag })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Back to all posts' })).toBeVisible();
  });

  test('shows untitled post on detail page', async ({ page }) => {
    await login(page);
    const content = uniqueId('untitled-body');

    await createPost(page, { content });

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await page
      .locator('.post')
      .filter({ hasText: content })
      .getByRole('link', { name: 'Untitled' })
      .click();

    await expect(page.locator('.post-title')).toContainText('Untitled');
    await expect(page.locator('.post-content')).toContainText(content);
  });

  test('returns 404 for non-existent post', async ({ page }) => {
    const response = await page.goto('/posts/999999999', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(404);
  });
});
