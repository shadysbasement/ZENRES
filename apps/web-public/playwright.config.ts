import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: 'tests',
	use: { baseURL: 'http://localhost:4173' },
	webServer: { command: 'npm run build && npm run preview', port: 4173, reuseExistingServer: true, env: { ADMIN_GATE_PATH: '/_v/e2e-gate' } },
	projects: [
		{ name: 'desktop-1440', use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } } },
		{ name: 'laptop-1024', use: { ...devices['Desktop Chrome'], viewport: { width: 1024, height: 768 } } },
		{ name: 'mobile-390', use: { ...devices['iPhone 13'] } }
	]
});
