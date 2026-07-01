import { expect, test } from '@playwright/test';

import { acceptNextDialog, login } from './helpers';

test.describe.serial('admin tools', () => {
  test('seeds posts via admin tools', async ({ page }) => {
    await login(page);
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await page.locator('#count').fill('3');
    await page.getByRole('button', { name: 'Seed Posts' }).click();

    await expect(page.locator('.alert-success')).toContainText('Created 3 post(s).');
  });

  test('executes SELECT query and shows results', async ({ page }) => {
    await login(page);
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await page.locator('#query').fill('SELECT id, title FROM posts ORDER BY id DESC LIMIT 5');
    await page.getByRole('button', { name: 'Execute' }).click();

    await expect(page.locator('body')).toContainText('row(s) returned');
    await expect(page.locator('table thead th')).toContainText(['id', 'title']);
    await expect(page.locator('table tbody tr').first()).toBeVisible();
  });

  test('rejects empty SQL query', async ({ page }) => {
    await login(page);
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await page.locator('#query').fill('');
    await page.getByRole('button', { name: 'Execute' }).click();

    await expect(page.locator('.error')).toContainText('Query is required');
  });

  test('rejects multiple SQL statements', async ({ page }) => {
    await login(page);
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await page.locator('#query').fill('SELECT 1; SELECT 2');
    await page.getByRole('button', { name: 'Execute' }).click();

    await expect(page.locator('.error')).toContainText('Only one SQL statement is allowed');
  });

  test('clears all posts and tags', async ({ page }) => {
    await login(page);
    await page.goto('/admin/tools', { waitUntil: 'domcontentloaded' });

    await acceptNextDialog(page);
    await page.getByRole('button', { name: 'Clear Posts' }).click();

    await expect(page.locator('.alert-success')).toContainText(/Deleted \d+ post\(s\) and \d+ tag\(s\)\./);

    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('.post')).toHaveCount(0);
    await expect(page.locator('body')).toContainText('No posts found.');
  });
});
