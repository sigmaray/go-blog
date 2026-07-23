import { expect, test } from '@playwright/test';

import {
  acceptNextDialog,
  createPost,
  login,
  setPostContent,
  uniqueId,
  waitForPostEditor,
} from './helpers';

test.describe.serial('posts', () => {
  test('validates required content on create', async ({ page }) => {
    await login(page);
    await page.goto('/admin/posts/new', { waitUntil: 'domcontentloaded' });
    await waitForPostEditor(page);

    await page.locator('#title').fill('Title without content');
    await page.getByRole('button', { name: /create post/i }).click();

    await expect(page.locator('.error')).toContainText('Content is required');
    await expect(page).toHaveURL(/\/admin\/posts$/);
    await expect(page.getByRole('heading', { name: 'Create New Post' })).toBeVisible();
  });

  test('preserves inline CSS styles in post HTML', async ({ page }) => {
    await login(page);
    const title = uniqueId('styled-post');
    const marker = uniqueId('styled-body');

    await createPost(page, {
      title,
      content: `<p style="color: red; font-size: 18px; text-align: center;">${marker}</p>`,
    });

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    const styled = page
      .locator('.post')
      .filter({ hasText: title })
      .locator('.post-content p')
      .filter({ hasText: marker });
    await expect(styled).toBeVisible();
    await expect(styled).toHaveCSS('color', 'rgb(255, 0, 0)');
    await expect(styled).toHaveCSS('font-size', '18px');
    await expect(styled).toHaveCSS('text-align', 'center');
  });

  test('creates a post with title, content, and tags', async ({ page }) => {
    await login(page);
    const title = uniqueId('feature-post');
    const firstTag = uniqueId('go');
    const secondTag = uniqueId('tutorial');

    await createPost(page, {
      title,
      content: 'Feature post body',
      tags: `${firstTag}, ${secondTag}`,
    });

    await expect(page.locator('body')).toContainText(title);

    await page.goto(`/?tag=${firstTag}`, { waitUntil: 'domcontentloaded' });
    const post = page.locator('.post').filter({ hasText: title });
    await expect(post).toBeVisible();
    await expect(post.locator('.post-content')).toContainText('Feature post body');
    await expect(post.getByRole('link', { name: firstTag })).toBeVisible();
    await expect(post.getByRole('link', { name: secondTag })).toBeVisible();
  });

  test('filters posts by tag', async ({ page }) => {
    await login(page);
    const tagGo = uniqueId('e2e-go');
    const tagWeb = uniqueId('e2e-web');
    const goTitle = uniqueId('go-tagged-post');
    const webTitle = uniqueId('web-tagged-post');

    await createPost(page, {
      title: goTitle,
      content: 'Content for go tag',
      tags: tagGo,
    });
    await createPost(page, {
      title: webTitle,
      content: 'Content for web tag',
      tags: tagWeb,
    });

    await page.goto(`/?tag=${tagGo}`, { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(`Showing posts for tag: ${tagGo}`);
    await expect(page.locator('.post')).toHaveCount(1);
    await expect(page.locator('.post')).toContainText(goTitle);
    await expect(page.locator('.post')).not.toContainText(webTitle);

    await page.getByRole('link', { name: 'Clear filter' }).click();
    await expect(page.locator('body')).not.toContainText('Showing posts for tag:');
    await expect(page.locator('.post').filter({ hasText: goTitle })).toBeVisible();
    await expect(page.locator('.post').filter({ hasText: webTitle })).toBeVisible();

    const goPost = page.locator('.post').filter({ hasText: goTitle });
    await goPost.getByRole('link', { name: tagGo }).click();
    await expect(page).toHaveURL(new RegExp(`[?&]tag=${tagGo}(?:&|$)`));
    await expect(page.locator('.post')).toHaveCount(1);
    await expect(page.locator('.post')).toContainText(goTitle);
  });

  test('edits a post', async ({ page }) => {
    await login(page);
    const title = uniqueId('edit-post');
    const updatedTitle = uniqueId('edit-post-updated');
    const tag = uniqueId('edit-tag');
    const updatedTag = uniqueId('edit-tag-updated');

    await createPost(page, {
      title,
      content: 'Original content',
      tags: tag,
    });

    const row = page.locator('tr').filter({ hasText: title });
    await row.getByRole('link', { name: 'Edit' }).click();
    await expect(page.getByRole('heading', { name: 'Edit Post' })).toBeVisible();

    await page.locator('#title').fill(updatedTitle);
    await setPostContent(page, 'Updated content');
    await page.locator('#tags').fill(updatedTag);
    await page.getByRole('button', { name: /update post/i }).click();
    await expect(page).toHaveURL(/\/admin\/?$/);
    await expect(page.locator('body')).toContainText(updatedTitle);
    await expect(page.locator('body')).not.toContainText(title);

    await page.goto(`/?tag=${updatedTag}`, { waitUntil: 'domcontentloaded' });
    const post = page.locator('.post').filter({ hasText: updatedTitle });
    await expect(post).toBeVisible();
    await expect(post.locator('.post-content')).toContainText('Updated content');
  });

  test('creates a post using markdown mode', async ({ page }) => {
    await login(page);
    const title = uniqueId('markdown-post');
    const marker = uniqueId('md-body');

    await page.goto('/admin/posts/new', { waitUntil: 'domcontentloaded' });
    await waitForPostEditor(page);
    await page.locator('#title').fill(title);
    await page.getByRole('button', { name: 'Markdown', exact: true }).click();
    await page.locator('#post-editor-markdown').fill(`**${marker}**`);
    await page.getByRole('button', { name: /create post/i }).click();
    await expect(page).toHaveURL(/\/admin\/?$/);

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    const post = page.locator('.post').filter({ hasText: title });
    await expect(post.locator('.post-content strong')).toContainText(marker);
  });

  test('creates a post using visual mode', async ({ page }) => {
    await login(page);
    const title = uniqueId('visual-post');
    const marker = uniqueId('visual-body');

    await page.goto('/admin/posts/new', { waitUntil: 'domcontentloaded' });
    await waitForPostEditor(page);
    await page.locator('#title').fill(title);
    await page.getByRole('button', { name: 'Visual', exact: true }).click();
    const visualFrame = page.frameLocator('iframe.tox-edit-area__iframe');
    await visualFrame.locator('body').click();
    await visualFrame.locator('body').fill(marker);
    await page.getByRole('button', { name: /create post/i }).click();
    await expect(page).toHaveURL(/\/admin\/?$/);

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    const post = page.locator('.post').filter({ hasText: title });
    await expect(post.locator('.post-content')).toContainText(marker);
  });

  test('switches between visual, markdown, and html modes', async ({ page }) => {
    await login(page);
    await page.goto('/admin/posts/new', { waitUntil: 'domcontentloaded' });
    await waitForPostEditor(page);

    await setPostContent(page, '<p>Mode switch sample</p>');
    await page.getByRole('button', { name: 'Markdown', exact: true }).click();
    await expect(page.locator('#post-editor-markdown')).toBeVisible();
    await expect(page.locator('#post-editor-markdown')).toHaveValue(/Mode switch sample/);

    await page.getByRole('button', { name: 'Visual', exact: true }).click();
    await expect(page.locator('iframe.tox-edit-area__iframe')).toBeVisible();

    await page.getByRole('button', { name: 'HTML', exact: true }).click();
    await expect(page.locator('#content')).toBeVisible();
    await expect(page.locator('#content')).toHaveValue(/Mode switch sample/);
  });

  test('deletes a post', async ({ page }) => {
    await login(page);
    const title = uniqueId('delete-post');

    await createPost(page, {
      title,
      content: 'Post to delete',
      tags: uniqueId('delete-tag'),
    });

    await acceptNextDialog(page);
    const row = page.locator('tr').filter({ hasText: title });
    await row.getByRole('button', { name: 'Delete' }).click();
    await expect(page).toHaveURL(/\/admin\/?$/);
    await expect(page.locator('body')).not.toContainText(title);

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('.post').filter({ hasText: title })).toHaveCount(0);
  });
});
