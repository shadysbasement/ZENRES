import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/public';
import { OPS_COOKIE, revokeOpsSession } from '$lib/server/opsVault';

export const GET: RequestHandler = async ({ cookies }) => {
	const token = cookies.get(OPS_COOKIE);
	if (token) {
		try {
			await revokeOpsSession(token);
		} catch {
			// Always clear the browser cookie, including when the auth API is unavailable.
		}
	}
	cookies.delete(OPS_COOKIE, { path: '/', secure: true });
	return new Response(null, {
		status: 303,
		headers: {
			location: env.PUBLIC_SITE_URL ?? 'http://localhost:5173',
			'cache-control': 'no-store',
			'referrer-policy': 'no-referrer'
		}
	});
};
