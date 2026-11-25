import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/public';
export const GET: RequestHandler = ({ cookies }) => {
	cookies.delete('zv_session', { path: '/' });
	return new Response(null, { status: 303, headers: { location: `${env.PUBLIC_SITE_URL ?? 'http://localhost:5173'}/login` } });
};
