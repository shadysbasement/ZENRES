import type { RequestHandler } from './$types';
import { OPS_COOKIE, cookieOpts, exchangeHandoff } from '$lib/server/opsVault';

/** One-time code from the PGP vault on zenres.example → host-only __Host-zv_ops cookie on app.zenres.example. */
export const GET: RequestHandler = async ({ url, cookies, getClientAddress, request }) => {
	const code = url.searchParams.get('code') ?? '';
	const notFound = () => new Response('Not Found', { status: 404, headers: { 'cache-control': 'no-store', 'referrer-policy': 'no-referrer' } });
	if (!/^[A-Za-z0-9_-]{32,64}$/.test(code)) return notFound();
	try {
		const r = await exchangeHandoff(code, getClientAddress(), request.headers.get('user-agent') ?? '');
		if (!r) return notFound();
		cookies.set(OPS_COOKIE, r.token, cookieOpts);
		return new Response(null, { status: 303, headers: { location: '/ops-vault', 'cache-control': 'no-store', 'referrer-policy': 'no-referrer' } });
	} catch {
		return notFound();
	}
};
