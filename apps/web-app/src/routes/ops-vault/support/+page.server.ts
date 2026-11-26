import { error, fail } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { Actions, PageServerLoad } from './$types';
import { OPS_COOKIE, clientIP } from '$lib/server/opsVault';

type Ticket = { id: string; number: number; requester: string; subject: string; category: string; kind: string; priority: string; status: string; plan: string; assigned_to: string; callback_at?: string; created_at: string; updated_at: string; first_response_at?: string; resolved_at?: string; rating?: number };
type Message = { id: string; author: string; body: string; internal: boolean; created_at: string };

async function supportFetch(event: { cookies: { get: (name: string) => string | undefined }; request: Request; getClientAddress: () => string; fetch: typeof fetch }, path: string, init: RequestInit = {}) {
	const base = env.AUTH_API_URL?.replace(/\/$/, '');
	if (!base) throw Error('support_unavailable');
	const token = event.cookies.get(OPS_COOKIE);
	if (!token) error(404, 'Not Found');
	return event.fetch(`${base}/internal/support/${path}`, {
		...init,
		headers: {
			'content-type': 'application/json',
			'x-internal-secret': env.INTERNAL_API_SECRET ?? '',
			'x-client-ip': clientIP(event.request, event.getClientAddress),
			'x-client-ua': event.request.headers.get('user-agent') ?? '',
			authorization: `Bearer ${token}`
		},
		signal: AbortSignal.timeout(9000)
	});
}

export const load: PageServerLoad = async (event) => {
	event.setHeaders({ 'cache-control': 'no-store' });
	const { admin } = await event.parent();
	let tickets: Ticket[] = [];
	let agents: { id: string; handle: string; role: string }[] = [];
	let costs: { customer: string; currency: string; tickets: number; minutes: number; amount: string }[] = [];
	let detail: { ticket: Ticket; messages: Message[]; rating: number | null } | null = null;
	let unavailable = false;
	try {
		const ticketResponse = await supportFetch(event, 'tickets');
		if (!ticketResponse.ok) throw Error('support_unavailable');
		tickets = await ticketResponse.json();
		if (admin.role === 'super_admin' || admin.role === 'support_admin') {
			const agentResponse = await supportFetch(event, 'agents');
			if (agentResponse.ok) agents = await agentResponse.json();
			const costResponse = await supportFetch(event, 'costs');
			if (costResponse.ok) costs = await costResponse.json();
		}
		const id = event.url.searchParams.get('id');
		if (id && /^[0-9a-f-]{36}$/i.test(id)) {
			const response = await supportFetch(event, `tickets/${id}`);
			if (response.ok) detail = await response.json();
		}
	} catch { unavailable = true; }
	return { tickets, agents, costs, detail, unavailable, canManage: admin.role === 'super_admin' || admin.role === 'support_admin' };
};

export const actions: Actions = {
	update: async (event) => {
		const form = await event.request.formData();
		const id = String(form.get('id') ?? '');
		if (!/^[0-9a-f-]{36}$/i.test(id)) return fail(400, { message: 'Invalid ticket.' });
		const body = {
			status: String(form.get('status') ?? ''), priority: String(form.get('priority') ?? ''),
			assignee_id: String(form.get('assignee_id') ?? ''), message: String(form.get('message') ?? ''),
			internal_note: form.get('internal_note') === 'on',
			work_minutes: Number(form.get('work_minutes') ?? 0), work_cost: String(form.get('work_cost') ?? ''),
			work_currency: String(form.get('work_currency') ?? ''), work_note: String(form.get('work_note') ?? '')
		};
		try {
			const response = await supportFetch(event, `tickets/${id}`, { method: 'POST', body: JSON.stringify(body) });
			if (!response.ok) return fail(response.status, { message: 'Update was not saved.' });
			return { message: 'Ticket updated.' };
		} catch { return fail(503, { message: 'Support service unavailable.' }); }
	},
	announcement: async (event) => {
		const form = await event.request.formData();
		const body = { title: String(form.get('title') ?? ''), body: String(form.get('body') ?? ''), state: String(form.get('state') ?? ''), ends_at: '' };
		try {
			const response = await supportFetch(event, 'announcements', { method: 'POST', body: JSON.stringify(body) });
			if (!response.ok) return fail(response.status, { message: 'Notice was not published.' });
			return { message: 'Notice published.' };
		} catch { return fail(503, { message: 'Support service unavailable.' }); }
	}
};
