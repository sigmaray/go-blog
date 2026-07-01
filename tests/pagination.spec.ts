import { expect, test } from '@playwright/test';

import { createTaggedPosts, login, pagination, uniqueId } from './helpers';

test.describe.serial('pagination', () => {
  test('shows page numbers with the current page highlighted', async ({ page }) => {
    await login(page);
    const tag = uniqueId('e2e-page-numbers');

    await createTaggedPosts(page, 6, { tag, titlePrefix: `${tag} post` });
    await page.goto(`/?tag=${tag}`, { waitUntil: 'domcontentloaded' });

    const nav = pagination(page);
    await expect(nav).toBeVisible();
    await expect(nav.getByRole('link', { name: '1' })).toHaveCount(0);
    await expect(nav.getByRole('link', { name: '2' })).toBeVisible();
    await expect(nav.locator('.page-item.active .page-link')).toHaveText('1');
    await expect(page.locator('.post')).toHaveCount(5);
  });

  test('navigates with Previous, Next, and page number links', async ({ page }) => {
    await login(page);
    const tag = uniqueId('e2e-page-nav');

    await createTaggedPosts(page, 6, { tag, titlePrefix: `${tag} post` });
    await page.goto(`/?tag=${tag}`, { waitUntil: 'domcontentloaded' });

    const nav = pagination(page);
    await expect(nav.getByRole('link', { name: 'Next' })).toBeVisible();
    await expect(nav.getByRole('link', { name: 'Previous' })).not.toBeVisible();

    await nav.getByRole('link', { name: '2' }).click();
    await expect(page).toHaveURL(
      new RegExp(`[?&]page=2(?:&|$).*tag=${tag}|[?&]tag=${tag}(?:&|$).*page=2`),
    );
    await expect(page.locator('.post')).toHaveCount(1);
    await expect(nav.locator('.page-item.active .page-link')).toHaveText('2');
    await expect(nav.getByRole('link', { name: '1' })).toBeVisible();
    await expect(nav.getByRole('link', { name: 'Previous' })).toBeVisible();
    await expect(nav.getByRole('link', { name: 'Next' })).not.toBeVisible();

    await nav.getByRole('link', { name: '1' }).click();
    await expect(page).toHaveURL(
      new RegExp(`[?&]page=1(?:&|$).*tag=${tag}|[?&]tag=${tag}(?:&|$).*page=1`),
    );
    await expect(page.locator('.post')).toHaveCount(5);
    await expect(nav.locator('.page-item.active .page-link')).toHaveText('1');

    await nav.getByRole('link', { name: 'Next' }).click();
    await expect(page).toHaveURL(
      new RegExp(`[?&]page=2(?:&|$).*tag=${tag}|[?&]tag=${tag}(?:&|$).*page=2`),
    );
    await expect(nav.locator('.page-item.active .page-link')).toHaveText('2');

    await nav.getByRole('link', { name: 'Previous' }).click();
    await expect(page.locator('.post')).toHaveCount(5);
    await expect(nav.locator('.page-item.active .page-link')).toHaveText('1');
  });

  test('keeps the tag filter in page number links', async ({ page }) => {
    await login(page);
    const tag = uniqueId('e2e-page-tag');

    await createTaggedPosts(page, 6, { tag, titlePrefix: `${tag} post` });
    await page.goto(`/?tag=${tag}`, { waitUntil: 'domcontentloaded' });

    const page2Link = pagination(page).getByRole('link', { name: '2' });
    await expect(page2Link).toHaveAttribute('href', new RegExp(`tag=${tag}`));
    await expect(page2Link).toHaveAttribute('href', /page=2/);
  });

  test('hides pagination when all posts fit on one page', async ({ page }) => {
    await login(page);
    const tag = uniqueId('e2e-single-page');

    await createTaggedPosts(page, 3, { tag, titlePrefix: `${tag} post` });
    await page.goto(`/?tag=${tag}`, { waitUntil: 'domcontentloaded' });

    await expect(page.locator('.post')).toHaveCount(3);
    await expect(pagination(page)).not.toBeVisible();
  });
});
