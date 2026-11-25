import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: 'tests',
	use: { baseURL: 'http://localhost:4174', ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
	webServer: { command: 'npm run build && npm run preview', port: 4174, reuseExistingServer: true }
});
