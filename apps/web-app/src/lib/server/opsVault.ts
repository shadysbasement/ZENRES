import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';

const AUTH_API = env.AUTH_API_URL ?? 'http://localhost:8081';
/** `__Host-` requires Secure; browsers treat http://localhost as a secure context, so this works in dev too. */
export const OPS_COOKIE = '__Host-zv_ops';

export const cookieOpts = {
	path: '/',
	httpOnly: true,
	secure: true,
	sameSite: 'lax' as const,
	maxAge: 4 * 60 * 60
};

/** Server-to-server check of the PGP-verified admin session (bound to UA+IP hash on the auth service). */
export async function validateOpsSession(token: string | undefined, ip: string, ua: string) {
	if (!token || !/^[A-Za-z0-9_-]{43,64}$/.test(token)) return null;
	try {
		const res = await fetch(`${AUTH_API}/internal/vault-auth/session`, {
			method: 'POST',
			headers: { 'content-type': 'application/json', 'x-internal-secret': env.INTERNAL_API_SECRET ?? '' },
			body: JSON.stringify({ token, ip, ua }),
			signal: AbortSignal.timeout(2000)
		});
		if (!res.ok) return null;
		return (await res.json()) as { handle: string; fingerprint: string; expires_at: string };
	} catch {
		if (dev) console.warn('[ops-vault] auth service unreachable — vault stays hidden');
		return null;
	}
}

export async function exchangeHandoff(code: string, ip: string, ua: string) {
	const res = await fetch(`${AUTH_API}/internal/vault-auth/exchange`, {
		method: 'POST',
		headers: { 'content-type': 'application/json', 'x-internal-secret': env.INTERNAL_API_SECRET ?? '' },
		body: JSON.stringify({ code, ip, ua }),
		signal: AbortSignal.timeout(3000)
	});
	if (!res.ok) return null;
	return (await res.json()) as { token: string };
}
