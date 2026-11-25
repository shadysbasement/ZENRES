import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({ out: 'build' }),
		alias: {
			$design: '../../packages/design/src'
		},
		// Inline scripts are hashed by SvelteKit; everything else is same-origin.
		csp: {
			mode: 'hash',
			directives: {
				'default-src': ['self'],
				'script-src': ['self'],
				'style-src': ['self', 'unsafe-inline'],
				'img-src': ['self', 'data:', 'blob:'],
				'font-src': ['self', 'data:'],
				'connect-src': ['self'],
				'frame-ancestors': ['none'],
				'form-action': ['self', 'https://accounts.google.com', 'https://appleid.apple.com'],
				'base-uri': ['self'],
				'object-src': ['none']
			}
		}
	}
};

export default config;
