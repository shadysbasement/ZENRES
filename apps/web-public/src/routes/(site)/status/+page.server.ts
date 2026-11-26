import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ fetch, setHeaders }) => {
	setHeaders({ 'cache-control': 'no-store' });
	const base = env.AUTH_API_URL?.trim();
	let auth: 'operational' | 'unavailable' | 'not configured' = 'not configured';
	if (base) {
		try {
			const response = await fetch(`${base.replace(/\/$/, '')}/healthz`, { signal: AbortSignal.timeout(2500), cache: 'no-store' });
			const body = response.ok ? await response.json() : null;
			auth = response.ok && body?.status === 'ok' ? 'operational' : 'unavailable';
		} catch {
			auth = 'unavailable';
		}
	}
	return { auth, checkedAt: new Date().toISOString() };
};
