import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

export const load: PageServerLoad = ({ locals, url }) => {
	// Auth required. In local demo mode (no auth service) the wizard stays reachable.
	if (!locals.user && env.PUBLIC_DEMO_MODE === 'false') {
		redirect(303, `/login?next=${encodeURIComponent(url.pathname + url.search)}`);
	}
	return { user: locals.user };
};
