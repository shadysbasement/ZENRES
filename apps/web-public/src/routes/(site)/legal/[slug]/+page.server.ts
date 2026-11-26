import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { LEGAL_POLICIES, type LegalSlug } from '$lib/legal/policies';
import { legalConfig } from '$lib/server/legal';

export const load: PageServerLoad = ({ params, setHeaders }) => {
	const policy = LEGAL_POLICIES[params.slug as LegalSlug];
	const config = legalConfig();
	if (!policy || !config.visible) error(404, 'Not Found');
	if (!config.published) setHeaders({ 'x-robots-tag': 'noindex, nofollow', 'cache-control': 'no-store' });
	return { policy, config, slug: params.slug };
};
