import { test, expect } from '@playwright/test';

test('admin-looking paths 404 identically', async ({ request }) => {
	for (const p of ['/admin', '/administrator', '/ops', '/pgp-login', '/wp-admin']) {
		const r = await request.get(p);
		expect(r.status(), p).toBe(404);
	}
});

test('robots.txt and sitemap.xml contain no vault reference', async ({ request }) => {
	const robots = await (await request.get('/robots.txt')).text();
	const sitemap = await (await request.get('/sitemap.xml')).text();
	for (const body of [robots, sitemap]) {
		expect(body).not.toContain('_v');
		expect(body.toLowerCase()).not.toContain('vault');
		expect(body.toLowerCase()).not.toContain('admin');
	}
});

test('gate route is noindex/no-store and renders the terminal', async ({ request, page }) => {
	const r = await request.get('/_v/e2e-gate');
	expect(r.status()).toBe(200);
	expect(r.headers()['x-robots-tag']).toContain('noindex');
	expect(r.headers()['cache-control']).toContain('no-store');
	await page.goto('/_v/e2e-gate');
	await expect(page.getByText('REQUEST ENCRYPTED CHALLENGE')).toBeVisible();
});

test('no client bundle references the gate path', async ({ page }) => {
	const scripts: string[] = [];
	page.on('response', async (res) => {
		if (res.url().endsWith('.js')) scripts.push(await res.text());
	});
	await page.goto('/');
	await page.waitForLoadState('networkidle');
	expect(scripts.join('\n')).not.toContain('e2e-gate');
});
