import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ fetch, setHeaders }) => {
	setHeaders({ 'cache-control': 'no-store' });
	let announcements: { id: string; title: string; body: string; state: string; published_at: string }[] = [];
	let announcementsAvailable = false;
	try {
		const res = await fetch('/api/support/announcements', { signal: AbortSignal.timeout(5000) });
		if (res.ok) { announcements = await res.json(); announcementsAvailable = true; }
	} catch { /* status notices are optional when the service is unavailable */ }
	const email = env.SUPPORT_EMAIL?.trim() ?? '';
	const whatsapp = env.SUPPORT_WHATSAPP?.trim() ?? '';
	return {
		announcements,
		announcementsAvailable,
		email: /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email) ? email : '',
		whatsapp: /^\+?[1-9]\d{7,14}$/.test(whatsapp) ? whatsapp.replace('+', '') : '',
		aiEnabled: !!(env.SUPPORT_AI_API_KEY && env.SUPPORT_AI_MODEL)
	};
};
