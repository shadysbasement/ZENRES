import type { Actions, PageServerLoad } from './$types';
import { error, fail, redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { gateSecretHeader, isGateRequest, rateLimit } from '$lib/server/gate';

/**
 * Catch-all route. Every unknown path 404s through here with identical behaviour, except the
 * single path configured in ADMIN_GATE_PATH, which renders the PGP vault terminal.
 * Because the matcher is a rest parameter, the client route manifest never contains the gate path.
 */
const AUTH_API = env.AUTH_API_URL ?? 'http://localhost:8081';

function notFound(): never {
	error(404, { message: 'Not Found' });
}

export const load: PageServerLoad = ({ url, setHeaders }) => {
	if (!isGateRequest(url.pathname)) notFound();
	setHeaders({ 'x-robots-tag': 'noindex, nofollow, noarchive', 'cache-control': 'no-store' });
	return { vault: true };
};

async function callVault(path: string, body: unknown, ip: string, ua: string) {
	const res = await fetch(`${AUTH_API}/api/v1/vault-auth/${path}`, {
		method: 'POST',
		headers: { 'content-type': 'application/json', 'x-forwarded-for': ip, 'user-agent': ua, ...gateSecretHeader() },
		body: JSON.stringify(body),
		signal: AbortSignal.timeout(8000)
	});
	const json = await res.json().catch(() => ({}));
	return { status: res.status, json: json as Record<string, unknown> };
}

export const actions: Actions = {
	challenge: async ({ request, url, getClientAddress }) => {
		if (!isGateRequest(url.pathname)) notFound();
		const ip = getClientAddress();
		if (!rateLimit(`vault-ch:${ip}`, 5, 15 * 60_000)) return fail(429, { stage: 1, error: 'RATE LIMITED — RETRY LATER' });
		const identifier = String((await request.formData()).get('identifier') ?? '')
			.replace(/\s+/g, '')
			.toUpperCase();
		if (!/^([0-9A-F]{40}|[0-9A-F]{64}|[A-Z0-9_.-]{3,40})$/.test(identifier)) return fail(400, { stage: 1, error: 'MALFORMED IDENTIFIER' });
		try {
			const { status, json } = await callVault('challenge', { identifier }, ip, request.headers.get('user-agent') ?? '');
			if (status !== 200) return fail(status === 429 ? 429 : 400, { stage: 1, error: status === 429 ? 'RATE LIMITED — RETRY LATER' : 'CHALLENGE UNAVAILABLE' });
			return { stage: 2, challenge_id: json.challenge_id as string, armored: json.armored_challenge as string, expires_in: Number(json.expires_in ?? 120), identifier };
		} catch {
			return fail(503, { stage: 1, error: 'VAULT SERVICE UNREACHABLE' });
		}
	},
	verify: async ({ request, url, getClientAddress }) => {
		if (!isGateRequest(url.pathname)) notFound();
		const ip = getClientAddress();
		const fd = await request.formData();
		const challenge_id = String(fd.get('challenge_id') ?? '');
		const signed = String(fd.get('signed') ?? '');
		const signature = String(fd.get('signature') ?? '');
		if (!challenge_id || !signed.includes('-----BEGIN PGP')) return fail(400, { stage: 2, error: 'PASTE THE CLEARSIGNED OUTPUT', challenge_id });
		let next: string | null = null;
		try {
			const { status, json } = await callVault('verify', { challenge_id, signed, signature: signature || undefined }, ip, request.headers.get('user-agent') ?? '');
			if (status !== 200) {
				const code = String(json.error ?? 'verify_failed').toUpperCase().replace(/_/g, ' ');
				return fail(401, { stage: status === 410 ? 1 : 2, error: `SIGNATURE REJECTED — ${code}`, challenge_id: status === 410 ? undefined : challenge_id });
			}
			next = String(json.handoff_url ?? '');
		} catch {
			return fail(503, { stage: 2, error: 'VAULT SERVICE UNREACHABLE', challenge_id });
		}
		if (next) redirect(303, next);
		return fail(500, { stage: 1, error: 'HANDOFF FAILED' });
	}
};
