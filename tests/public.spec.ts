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

  test('hides post from public index and detail until unhidden', async ({ page }) => {
    await login(page);
    const title = uniqueId('hidden-post');

    await createPost(page, {
      title,
      content: 'Hidden post body',
      hidden: true,
    });

    const row = page.locator('tr').filter({ hasText: title });
    await expect(row).toBeVisible();
    await expect(row.getByText('Hidden', { exact: true })).toBeVisible();

    const editHref = await row.getByRole('link', { name: 'Edit' }).getAttribute('href');
    expect(editHref).toMatch(/\/admin\/posts\/\d+\/edit/);
    const postId = editHref!.match(/\/admin\/posts\/(\d+)\/edit/)![1];

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('.post').filter({ hasText: title })).toHaveCount(0);

    const hiddenResponse = await page.goto(`/posts/${postId}`, { waitUntil: 'domcontentloaded' });
    expect(hiddenResponse?.status()).toBe(404);

    await page.goto(`/admin/posts/${postId}/edit`, { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#hidden')).toBeChecked();
    await page.locator('#hidden').uncheck();
    await page.getByRole('button', { name: /update post/i }).click();
    await expect(page).toHaveURL(/\/admin\/?$/);
    await expect(page.locator('tr').filter({ hasText: title }).getByText('Hidden', { exact: true })).toHaveCount(0);

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('.post').filter({ hasText: title })).toBeVisible();

    const visibleResponse = await page.goto(`/posts/${postId}`, { waitUntil: 'domcontentloaded' });
    expect(visibleResponse?.status()).toBe(200);
    await expect(page.locator('.post-title')).toContainText(title);
  });
});
