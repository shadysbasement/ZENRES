import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';
import { env as pub } from '$env/dynamic/public';
import { legalConfig } from '$lib/server/legal';

/**
 * Same-origin proxy to the Go auth service so session cookies stay first-party and the
 * browser never needs CORS. Only the public /api/v1/auth/* and onboarding surface is proxied.
 */
const AUTH_API = env.AUTH_API_URL ?? 'http://localhost:8081';
const ALLOWED = /^(register|login|logout|me|consent(?:\/marketing)?|verify-email|password\/(forgot|reset)|proton\/(magic-link|verify|wkd)|passkey\/(register|login)\/(start|finish)|oauth\/(google|apple)\/(start|callback)|onboarding\/complete|link\/confirm)$/;

const handler: RequestHandler = async ({ params, request, url, getClientAddress }) => {
	if (!ALLOWED.test(params.path)) return new Response('Not found', { status: 404 });
	const target = params.path === 'onboarding/complete' ? `${AUTH_API}/api/v1/onboarding/complete` : `${AUTH_API}/api/v1/auth/${params.path}${url.search}`;
	const headers = new Headers();
	for (const h of ['content-type', 'cookie', 'user-agent', 'x-csrf-token', 'accept']) {
		const v = request.headers.get(h);
		if (v) headers.set(h, v);
	}
	headers.set('x-forwarded-for', getClientAddress());
	headers.set('x-forwarded-host', url.host);
	headers.set('x-forwarded-proto', url.protocol.replace(':', ''));
	try {
		const res = await fetch(target, {
			method: request.method,
			headers,
			body: request.method === 'GET' || request.method === 'HEAD' ? undefined : await request.arrayBuffer(),
			redirect: 'manual',
			signal: AbortSignal.timeout(10_000)
		});
		const out = new Headers();
		res.headers.forEach((v, k) => {
			if (!['content-encoding', 'content-length', 'transfer-encoding', 'connection', 'set-cookie'].includes(k)) out.append(k, v);
		});
		for (const c of res.headers.getSetCookie?.() ?? []) out.append('set-cookie', c);
		return new Response(res.body, { status: res.status, headers: out });
	} catch {
		// Demo fallback: when the auth service is not running locally, OAuth starts continue into onboarding.
		if (!legalConfig().published && pub.PUBLIC_DEMO_MODE !== 'false' && /^oauth\/(google|apple)\/start$/.test(params.path)) {
			return new Response(null, { status: 303, headers: { location: `/onboarding?via=${params.path.split('/')[1]}` } });
		}
		return new Response(JSON.stringify({ error: 'auth_unavailable' }), { status: 503, headers: { 'content-type': 'application/json' } });
	}
};

export const GET = handler;
export const POST = handler;
