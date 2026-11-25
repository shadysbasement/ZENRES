import { error } from '@sveltejs/kit';
import { ws } from '$lib/stores/workspace.svelte';
import { moduleForPath } from '$lib/entitlements/manifest';
import type { LayoutLoad } from './$types';

// The authenticated workspace is a client-rendered SPA shell.
export const ssr = false;

let initialised = false;

export const load: LayoutLoad = ({ url }) => {
	if (!initialised) {
		ws.init(url);
		initialised = true;
	}
	// Zero-lock: modules outside the manifest simply do not exist for this tenant.
	const mod = moduleForPath(url.pathname);
	if (mod && !ws.manifest.visible_modules.includes(mod)) error(404, { message: 'Not Found' });
	return {};
};
