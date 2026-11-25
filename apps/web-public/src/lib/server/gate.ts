import { env } from '$env/dynamic/private';
import { createHash, timingSafeEqual } from 'node:crypto';
import { isIP } from 'node:net';

/** Vercel supplies the original caller in this trusted platform header. */
export function clientIP(request: Request, fallback: () => string): string {
	const forwarded = env.VERCEL ? request.headers.get('x-vercel-forwarded-for')?.split(',')[0]?.trim() : '';
	return forwarded && isIP(forwarded) ? forwarded : fallback();
}

/**
 * Hidden PGP admin gate configuration. The path lives ONLY in server env
 * (ADMIN_GATE_PATH) — it is never referenced in client bundles, sitemap or robots.
 */
export function gatePath(): string | null {
	const p = env.ADMIN_GATE_PATH?.trim();
	if (!p || p === '/' || !p.startsWith('/')) return null;
	return p.replace(/\/+$/, '');
}

export function isGateRequest(pathname: string): boolean {
	const p = gatePath();
	return !!p && pathname.replace(/\/+$/, '') === p;
}

/** Constant-time check of the stealth-trigger passphrase hash (sha256 hex). */
export function checkTriggerHash(candidateHex: string): boolean {
	const expected = env.ADMIN_GATE_PASSPHRASE_SHA256?.trim().toLowerCase();
	if (!expected || !/^[a-f0-9]{64}$/.test(expected) || !/^[a-f0-9]{64}$/.test(candidateHex)) {
		// burn comparable work to keep timing flat
		createHash('sha256').update(candidateHex).digest();
		return false;
	}
	return timingSafeEqual(Buffer.from(expected, 'hex'), Buffer.from(candidateHex, 'hex'));
}

export function gateSecretHeader(): Record<string, string> {
	return env.ADMIN_GATE_SECRET ? { 'X-Zevra-Gate': env.ADMIN_GATE_SECRET } : {};
}

/** Simple fixed-window limiter (per process). Production deployments front this with Redis. */
const windows = new Map<string, { count: number; reset: number }>();
export function rateLimit(key: string, max: number, windowMs: number): boolean {
	const now = Date.now();
	const w = windows.get(key);
	if (!w || w.reset < now) {
		windows.set(key, { count: 1, reset: now + windowMs });
		return true;
	}
	w.count++;
	return w.count <= max;
}
