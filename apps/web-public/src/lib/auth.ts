import { env } from '$env/dynamic/public';

export const DEMO = env.PUBLIC_DEMO_MODE !== 'false';
export const APP_URL = env.PUBLIC_APP_URL ?? 'http://localhost:5174';

export type ApiResult<T = Record<string, unknown>> = { ok: true; data: T } | { ok: false; status: number; error: string; offline?: boolean };

function csrf(): string {
	const m = document.cookie.match(/(?:^|; )zv_csrf=([^;]+)/);
	return m ? decodeURIComponent(m[1]) : '';
}

export async function api<T = Record<string, unknown>>(path: string, body?: unknown, method = 'POST'): Promise<ApiResult<T>> {
	try {
		const res = await fetch(`/api/auth/${path}`, {
			method,
			headers: { 'content-type': 'application/json', 'x-csrf-token': csrf() },
			body: body === undefined ? undefined : JSON.stringify(body),
			credentials: 'same-origin'
		});
		const json = await res.json().catch(() => ({}));
		if (res.ok) return { ok: true, data: json as T };
		return { ok: false, status: res.status, error: (json as { error?: string }).error ?? 'request_failed', offline: res.status === 503 };
	} catch {
		return { ok: false, status: 0, error: 'network', offline: true };
	}
}

export const PROTON_DOMAINS = ['proton.me', 'protonmail.com', 'pm.me', 'protonmail.ch'];
export const isProtonDomain = (email: string) => PROTON_DOMAINS.includes(email.split('@')[1]?.toLowerCase() ?? '');

export function passwordScore(pw: string): { score: 0 | 1 | 2 | 3 | 4; label: string } {
	if (!pw) return { score: 0, label: '' };
	let classes = 0;
	if (/[a-z]/.test(pw)) classes++;
	if (/[A-Z]/.test(pw)) classes++;
	if (/[0-9]/.test(pw)) classes++;
	if (/[^A-Za-z0-9]/.test(pw)) classes++;
	const unique = new Set(pw).size;
	const bits = pw.length * Math.log2(Math.max(2, [26, 52, 62, 94][classes - 1] ?? 26)) * Math.min(1, unique / Math.max(6, pw.length * 0.6));
	const score = (bits < 40 ? 1 : bits < 60 ? 2 : bits < 80 ? 3 : 4) as 1 | 2 | 3 | 4;
	return { score: pw.length < 12 ? (Math.min(score, 2) as 1 | 2) : score, label: ['', 'Weak', 'Fair', 'Strong', 'Excellent'][score] };
}

/* ---------- WebAuthn helpers (Proton Pass / device passkeys) ---------- */

const b64uToBuf = (s: string) => {
	const pad = '='.repeat((4 - (s.length % 4)) % 4);
	const bin = atob((s + pad).replace(/-/g, '+').replace(/_/g, '/'));
	return Uint8Array.from(bin, (c) => c.charCodeAt(0)).buffer;
};
const bufToB64u = (b: ArrayBuffer) =>
	btoa(String.fromCharCode(...new Uint8Array(b)))
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=+$/, '');

type CreationJSON = {
	publicKey: Omit<PublicKeyCredentialCreationOptions, 'challenge' | 'user' | 'excludeCredentials'> & {
		challenge: string;
		user: { id: string; name: string; displayName: string };
		excludeCredentials?: { id: string; type: 'public-key' }[];
	};
};
type RequestJSON = {
	publicKey: Omit<PublicKeyCredentialRequestOptions, 'challenge' | 'allowCredentials'> & { challenge: string; allowCredentials?: { id: string; type: 'public-key' }[] };
};

export async function enrollPasskey(): Promise<ApiResult> {
	const start = await api<CreationJSON>('passkey/register/start', {});
	if (!start.ok) return start;
	const pk = start.data.publicKey;
	const cred = (await navigator.credentials.create({
		publicKey: {
			...pk,
			challenge: b64uToBuf(pk.challenge),
			user: { ...pk.user, id: b64uToBuf(pk.user.id) },
			excludeCredentials: pk.excludeCredentials?.map((c) => ({ ...c, id: b64uToBuf(c.id) }))
		}
	})) as PublicKeyCredential | null;
	if (!cred) return { ok: false, status: 0, error: 'cancelled' };
	const r = cred.response as AuthenticatorAttestationResponse;
	return api('passkey/register/finish', {
		id: cred.id,
		rawId: bufToB64u(cred.rawId),
		type: cred.type,
		response: { clientDataJSON: bufToB64u(r.clientDataJSON), attestationObject: bufToB64u(r.attestationObject), transports: r.getTransports?.() ?? [] }
	});
}

export async function loginWithPasskey(email?: string): Promise<ApiResult> {
	const start = await api<RequestJSON>('passkey/login/start', { email });
	if (!start.ok) return start;
	const pk = start.data.publicKey;
	const cred = (await navigator.credentials.get({
		publicKey: { ...pk, challenge: b64uToBuf(pk.challenge), allowCredentials: pk.allowCredentials?.map((c) => ({ ...c, id: b64uToBuf(c.id) })) }
	})) as PublicKeyCredential | null;
	if (!cred) return { ok: false, status: 0, error: 'cancelled' };
	const r = cred.response as AuthenticatorAssertionResponse;
	return api('passkey/login/finish', {
		id: cred.id,
		rawId: bufToB64u(cred.rawId),
		type: cred.type,
		response: {
			clientDataJSON: bufToB64u(r.clientDataJSON),
			authenticatorData: bufToB64u(r.authenticatorData),
			signature: bufToB64u(r.signature),
			userHandle: r.userHandle ? bufToB64u(r.userHandle) : null
		}
	});
}
