import type { RequestHandler } from './$types';
import { legalConfig } from '$lib/server/legal';
import { LEGAL_LINKS } from '$lib/legal/policies';

const PATHS = ['/', '/pricing', '/freightlink', '/login', '/signup'];

export const GET: RequestHandler = ({ url }) => {
	const paths = legalConfig().published ? [...PATHS, ...LEGAL_LINKS.map(({ slug }) => `/legal/${slug}`)] : PATHS;
	const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${paths.map((p) => `  <url><loc>${url.origin}${p}</loc><changefreq>weekly</changefreq></url>`).join('\n')}
</urlset>`;
	return new Response(body, { headers: { 'content-type': 'application/xml', 'cache-control': 'public, max-age=3600' } });
};
