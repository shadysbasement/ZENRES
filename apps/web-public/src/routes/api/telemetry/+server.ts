import type { RequestHandler } from './$types';
import { checkTriggerHash, gatePath, rateLimit } from '$lib/server/gate';

/**
 * Generic client telemetry sink. Also resolves the stealth trigger: when the posted key hash
 * matches the configured passphrase hash, the response carries the gate location. Every
 * other request receives an identical empty 204.
 */
export const POST: RequestHandler = async ({ request, getClientAddress }) => {
	const ip = getClientAddress();
	const empty = () => new Response(null, { status: 204, headers: { 'cache-control': 'no-store' } });
	if (!rateLimit(`tele:${ip}`, 10, 15 * 60_000)) return empty();
	let body: unknown;
	try {
		body = await request.json();
	} catch {
		return empty();
	}
	const k = typeof (body as { k?: unknown })?.k === 'string' ? ((body as { k: string }).k as string).toLowerCase() : '';
	const target = gatePath();
	if (target && k && checkTriggerHash(k)) {
		return new Response(JSON.stringify({ r: target }), {
			status: 200,
			headers: { 'content-type': 'application/json', 'cache-control': 'no-store', 'x-robots-tag': 'noindex' }
		});
	}
	return empty();
};
