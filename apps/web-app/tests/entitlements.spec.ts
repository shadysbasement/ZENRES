import { test, expect } from '@playwright/test';

const asTier = (tier: string) => async ({ page }: { page: import('@playwright/test').Page }) => {
	await page.addInitScript((t) => {
		localStorage.setItem('zv-tier', t);
		localStorage.removeItem('zv-hint-dismissals');
		sessionStorage.setItem('zv-app-preloaded', '1');
	}, tier);
};

test.describe('Starter tier — silent gating', () => {
	test.beforeEach(asTier('starter'));

	test('no padlocks and no higher-tier modules in the DOM', async ({ page }) => {
		await page.goto('/dashboard');
		const html = await page.content();
		expect(html).not.toContain('/freight/');
		expect(html).not.toContain('/importer-hub');
		expect(html).not.toContain('/carbon/simulator');
		expect(await page.locator('[class*="lock"]').count()).toBe(0);
	});

	test('direct visit to a Growth module 404s', async ({ page }) => {
		await page.goto('/freight/marketplace');
		await expect(page.getByText('404')).toBeVisible();
	});

	test('typing "freight" in Cmd+K surfaces a dismissible whisper', async ({ page }) => {
		await page.goto('/dashboard');
		await page.keyboard.press('Meta+k');
		await page.getByPlaceholder(/Search pages/).fill('freight');
		const w = page.locator('.whisper');
		await expect(w).toContainText('FreightLink backhaul matching is included in Growth');
		await w.getByRole('button', { name: 'Dismiss suggestion' }).click();
		await expect(w).toHaveCount(0);
	});

	test('mapping a 6th SKU whispers instead of blocking', async ({ page }) => {
		await page.goto('/products');
		await page.getByRole('button', { name: 'Map new SKU' }).click();
		await page.getByLabel('SKU code').fill('ALT-3003');
		await page.getByLabel('Product name').fill('Aluminium tubes');
		await page.getByLabel(/EU CN code/).fill('76083000');
		await page.getByRole('button', { name: 'Map SKU' }).click();
		await page.getByRole('button', { name: 'Map new SKU' }).click();
		await expect(page.locator('.whisper')).toContainText('all 5 CN-code SKUs');
	});
});

test.describe('Growth tier — digital twin & bell', () => {
	test.beforeEach(asTier('growth'));

	test('WebGL canvas renders and the bell increments on inbound mail', async ({ page }) => {
		const errors: string[] = [];
		page.on('pageerror', (e) => errors.push(e.message));
		await page.goto('/dashboard');
		await expect(page.locator('canvas[data-engine^="three.js"]')).toBeVisible();
		// The bell shows a dot; its accessible name carries the unread count.
		const bell = page.getByRole('button', { name: /^Notifications, \d+ unread$/ });
		const unread = async () => Number((await bell.getAttribute('aria-label'))?.match(/\d+/)?.[0]);
		const before = await unread();
		await page.getByRole('button', { name: 'Simulate buyer email' }).click();
		await expect(page.getByRole('button', { name: `Notifications, ${before + 1} unread` })).toBeVisible();
		expect(errors).toEqual([]);
	});

	test('zone inspector shows seeded workers and SKUs', async ({ page }) => {
		await page.goto('/dashboard?zone=PROD-A');
		await expect(page.getByText('Line A · Smelting & Casting').last()).toBeVisible();
		await expect(page.getByText('A-104').last()).toBeVisible();
	});

	test('Tally CSV imported twice yields zero duplicates', async ({ page }) => {
		await page.goto('/integrations/tally');
		await page.getByRole('button', { name: 'Use sample Day Book' }).click();
		await page.getByRole('button', { name: 'Import with deduplication' }).click();
		await expect(page.getByText('new vouchers')).toBeVisible();
		await page.getByRole('button', { name: 'Re-import same file' }).click();
		await expect(page.locator('.res > div').first()).toContainText('0');
	});
});

test('ops vault is 404 without a PGP-verified session', async ({ request }) => {
	expect((await request.get('/ops-vault')).status()).toBe(404);
});
