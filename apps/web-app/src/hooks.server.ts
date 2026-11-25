import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.opsAdmin = null;
	const response = await resolve(event);
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Permissions-Policy', 'camera=(self), microphone=(), geolocation=(self)');
	if (event.url.protocol === 'https:') response.headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains; preload');
	if (event.url.pathname.startsWith('/ops-vault')) {
		response.headers.set('X-Robots-Tag', 'noindex, nofollow, noarchive');
		response.headers.set('Cache-Control', 'no-store');
	}
	return response;
};
