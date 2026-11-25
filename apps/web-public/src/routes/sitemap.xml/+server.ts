import type { RequestHandler } from './$types';

const PATHS = ['/', '/pricing', '/freightlink', '/login', '/signup'];

export const GET: RequestHandler = ({ url }) => {
	const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${PATHS.map((p) => `  <url><loc>${url.origin}${p}</loc><changefreq>weekly</changefreq></url>`).join('\n')}
</urlset>`;
	return new Response(body, { headers: { 'content-type': 'application/xml', 'cache-control': 'public, max-age=3600' } });
};
