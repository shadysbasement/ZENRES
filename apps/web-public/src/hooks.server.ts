import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { isGateRequest } from '$lib/server/gate';

const AUTH_API = env.AUTH_API_URL ?? 'http://localhost:8081';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = null;
	const session = event.cookies.get('zv_session');
	if (session) {
		try {
			const res = await fetch(`${AUTH_API}/api/v1/auth/me`, {
				headers: { cookie: `zv_session=${session}` },
				signal: AbortSignal.timeout(1500)
			});
			if (res.ok) event.locals.user = await res.json();
		} catch {
			/* auth service unreachable — treat as signed out */
		}
	}

	const response = await resolve(event, {
		preload: ({ type }) => type === 'font' || type === 'js' || type === 'css'
	});

	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=(), publickey-credentials-get=(self), publickey-credentials-create=(self)');
	if (event.url.protocol === 'https:') {
		response.headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains; preload');
	}
	if (isGateRequest(event.url.pathname)) {
		response.headers.set('X-Robots-Tag', 'noindex, nofollow, noarchive');
		response.headers.set('Cache-Control', 'no-store');
	}
	return response;
};
