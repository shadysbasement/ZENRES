import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { env as pub } from '$env/dynamic/public';
import { PASSPORTS, type PublicBatch, type PublicContainer } from '$lib/server/passports';

const CORE_API = env.CORE_API_URL ?? 'http://localhost:8082';

export const load: PageServerLoad = async ({ params, setHeaders, fetch }) => {
	setHeaders({ 'cache-control': 'public, max-age=60' });
	const token = params.token.toUpperCase();
	if (!/^[0-9A-Z]{26}$/.test(token)) return { state: 'not_found' as const, type: params.type };

	try {
		const res = await fetch(`${CORE_API}/api/v1/public/verify/${params.type}/${token}`, { signal: AbortSignal.timeout(2000) });
		if (res.status === 404) return { state: 'not_found' as const, type: params.type };
		if (res.ok) {
			const passport = (await res.json()) as PublicBatch | PublicContainer;
			return { state: passport.status === 'revoked' ? ('revoked' as const) : ('ok' as const), type: params.type, passport, source: 'live' };
		}
	} catch {
		/* fall through to demo fixtures */
	}

	if (pub.PUBLIC_DEMO_MODE !== 'false') {
		const passport = PASSPORTS[token];
		if (!passport || passport.type !== params.type) return { state: 'not_found' as const, type: params.type };
		return { state: passport.status === 'revoked' ? ('revoked' as const) : ('ok' as const), type: params.type, passport, source: 'demo' };
	}
	return { state: 'unavailable' as const, type: params.type };
};
