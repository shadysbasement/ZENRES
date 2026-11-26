import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { legalConfig } from '$lib/server/legal';

export const load: PageServerLoad = ({ setHeaders }) => {
	const legal = legalConfig();
	if (!legal.visible) error(404, 'Not Found');
	if (!legal.published) setHeaders({ 'x-robots-tag': 'noindex, nofollow', 'cache-control': 'no-store' });
	return { legalPublished: legal.published };
};
