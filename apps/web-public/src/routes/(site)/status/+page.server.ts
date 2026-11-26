import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ fetch, setHeaders }) => {
	setHeaders({ 'cache-control': 'no-store' });
	const base = env.AUTH_API_URL?.trim();
	let auth: 'operational' | 'unavailable' | 'not configured' = 'not configured';
	if (base) {
		try {
			const response = await fetch(`${base.replace(/\/$/, '')}/healthz`, { signal: AbortSignal.timeout(6000), cache: 'no-store' });
			const body = response.ok ? await response.json() : null;
			auth = response.ok && body?.status === 'ok' ? 'operational' : 'unavailable';
		} catch {
			auth = 'unavailable';
		}
	}
	let announcements: { id: string; title: string; body: string; state: string; published_at: string }[] = [];
	let noticesAvailable = false;
	try {
		const response = await fetch('/api/support/announcements', { signal: AbortSignal.timeout(5000), cache: 'no-store' });
		if (response.ok) { announcements = await response.json(); noticesAvailable = true; }
	} catch { /* notices are shown only when the support service responds */ }
	return { auth, checkedAt: new Date().toISOString(), announcements, noticesAvailable };
};
