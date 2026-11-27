import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
	await page.addInitScript(() => sessionStorage.setItem('zv-preloaded', '1'));
});

test('landing renders hero, pillars, pricing and disclaimers without console errors', async ({ page }) => {
	const errors: string[] = [];
	page.on('pageerror', (e) => errors.push(e.message));
	await page.goto('/');
	await expect(page.getByRole('heading', { level: 1 })).toContainText('EU CBAM proof');
	await expect(page.getByText('One ledger.')).toBeVisible();
	await expect(page.getByText('Illustrative estimate').first()).toBeVisible();
	await expect(page.getByText('not legal, tax or accounting advice').first()).toBeAttached();
	// The hero is a live 3D scene, so assert its structure rather than a pixel snapshot.
	await expect(page.getByRole('tablist', { name: 'Explore the platform' })).toBeVisible();
	await expect(page.locator('.stage canvas, .stage .fallback').first()).toBeAttached({ timeout: 15_000 });
	await expect(page.getByRole('heading', { name: /From a Tally voucher/ })).toBeAttached();
	expect(errors).toEqual([]);
});

test('pricing toggles: monthly/annual and USD/INR/EUR', async ({ page }) => {
	await page.goto('/pricing');
	const growth = page.locator('#growth');
	await expect(growth).toContainText('$249');
	await page.getByRole('button', { name: /Annual/ }).click();
	await expect(growth).toContainText('$2,490');
	await page.getByRole('button', { name: 'INR' }).click();
	await expect(growth).toContainText('₹');
	await expect(page.getByText('14 dimensions')).toBeVisible();
});

test('FreightLink corridor steps update the readout', async ({ page }) => {
	await page.goto('/freightlink');
	await page.getByRole('button', { name: /Corridor match/ }).click();
	await expect(page.getByText('Hard filters first').first()).toBeVisible();
	await expect(page.getByText('40 L').first()).toBeVisible();
	await page.getByRole('button', { name: 'Ludhiana ↔ Delhi' }).click();
	await expect(page.getByText('310 km')).toBeVisible();
});

test('QR resolver: valid, revoked and unknown tokens', async ({ page }) => {
	await page.goto('/verify/c/01JD8ZV001CNTR7Q4M2X9KPWBA');
	await expect(page.getByText('ZV-001').first()).toBeVisible();
	await expect(page.getByText('Partner authentication required')).toBeVisible();
	await page.goto('/verify/p/01JD8ZA099BTCHREVOKED0000X');
	await expect(page.getByText('Revoked label')).toBeVisible();
	await page.goto('/verify/p/01JD8ZZZZZZZZZZZZZZZZZZZZZ');
	await expect(page.getByText('could not be resolved')).toBeVisible();
});
