import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';

const allowed = /^(announcements|tickets(?:\/[0-9a-f-]{36}(?:\/(?:messages|feedback))?)?)$/;
const handler: RequestHandler = async ({ params, request, url, getClientAddress }) => {
	if (!allowed.test(params.path)) return new Response('Not found', { status: 404 });
	const base = env.AUTH_API_URL?.replace(/\/$/, '');
	if (!base) return Response.json({ error: 'support_unavailable' }, { status: 503 });
	const headers = new Headers();
	for (const name of ['content-type', 'cookie', 'user-agent', 'x-csrf-token', 'accept']) {
		const value = request.headers.get(name);
		if (value) headers.set(name, value);
	}
	headers.set('x-forwarded-for', getClientAddress());
	try {
		const response = await fetch(`${base}/api/v1/support/${params.path}${url.search}`, {
			method: request.method,
			headers,
			body: request.method === 'GET' ? undefined : await request.arrayBuffer(),
			redirect: 'manual',
			signal: AbortSignal.timeout(10000)
		});
		const outgoing = new Headers(response.headers);
		for (const name of ['content-encoding', 'content-length', 'transfer-encoding', 'connection', 'set-cookie']) outgoing.delete(name);
		outgoing.set('cache-control', 'no-store');
		return new Response(response.body, { status: response.status, headers: outgoing });
	} catch {
		return Response.json({ error: 'support_unavailable' }, { status: 503 });
	}
};

export const GET = handler;
export const POST = handler;
