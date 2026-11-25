import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { OPS_COOKIE, validateOpsSession } from '$lib/server/opsVault';

/** Strictly unreachable without a valid PGP-verified admin session — otherwise a plain 404. */
export const load: LayoutServerLoad = async ({ cookies, getClientAddress, request, setHeaders }) => {
	const session = await validateOpsSession(cookies.get(OPS_COOKIE), getClientAddress(), request.headers.get('user-agent') ?? '');
	if (!session) error(404, { message: 'Not Found' });
	setHeaders({ 'cache-control': 'no-store', 'x-robots-tag': 'noindex, nofollow, noarchive' });
	return { admin: session };
};
