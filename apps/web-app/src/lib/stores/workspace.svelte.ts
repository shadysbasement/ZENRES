import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';
import { buildManifest, TIER_LABEL, type SubscriptionTier } from '$lib/entitlements/manifest';
import { evaluateSubtleHint, type HintTrigger, type SubtleHint } from '$lib/entitlements/hintEngine';
import * as seed from '$lib/demo/seed';
import type { Ccy } from '$lib/format';

const CORE_API = env.PUBLIC_CORE_API_URL ?? '';
const clone = <T>(v: T): T => structuredClone(v);

type Toast = { id: number; text: string; tone?: 'ok' | 'warn' | 'err' };

class Workspace {
	tier = $state<SubscriptionTier>('growth');
	manifest = $derived(buildManifest(this.tier));
	tierLabel = $derived(TIER_LABEL[this.tier]);
	orgName = $state(seed.ORG.name);
	mode = $state<'demo' | 'live'>('demo');
	currency = $state<Ccy>('USD');

	plants = $state(clone(seed.PLANTS));
	zones = $state(clone(seed.ZONES));
	products = $state(clone(seed.PRODUCTS));
	batches = $state(clone(seed.BATCHES));
	containers = $state(clone(seed.CONTAINERS));
	inventory = $state(clone(seed.INVENTORY));
	movements = $state(clone(seed.MOVEMENTS));
	vouchers = $state(clone(seed.VOUCHERS));
	notifications = $state(clone(seed.NOTIFICATIONS));
	freight = $state(clone(seed.FREIGHT));
	connect = $state(clone(seed.CONNECT));
	suppliers = $state(clone(seed.IMPORTER_SUPPLIERS));
	team = $state(clone(seed.TEAM));
	activity = $state(clone(seed.ACTIVITY));

	unread = $derived(this.notifications.filter((n) => !n.read).length);
	totalWorkers = $derived(this.zones.reduce((s, z) => s + z.active_workers, 0));
	totalTarget = $derived(this.zones.reduce((s, z) => s + z.shift_target_workers, 0));

	selectedZone = $state<string | null>(null);
	bellOpen = $state(false);
	paletteOpen = $state(false);
	toasts = $state<Toast[]>([]);
	/** Active whispers keyed by UI context (e.g. 'plants', 'palette'). */
	hints = $state<Record<string, SubtleHint | null>>({});
	dismissals = $state<Record<string, number>>({});

	#channel: BroadcastChannel | null = null;
	#tick: ReturnType<typeof setInterval> | null = null;
	#toastId = 0;

	init(url: URL) {
		if (!browser) return;
		const qTier = url.searchParams.get('demo_tier') as SubscriptionTier | null;
		const stored = localStorage.getItem('zv-tier') as SubscriptionTier | null;
		const t = qTier ?? stored;
		if (t && ['starter', 'growth', 'importer_hub'].includes(t)) this.setTier(t);
		const org = url.searchParams.get('org');
		if (org) {
			this.orgName = org;
			localStorage.setItem('zv-org', org);
		} else {
			const o = localStorage.getItem('zv-org');
			if (o) this.orgName = o;
		}
		try {
			this.dismissals = JSON.parse(localStorage.getItem('zv-hint-dismissals') ?? '{}');
		} catch {
			this.dismissals = {};
		}
		this.currency = (localStorage.getItem('zv-ccy') as Ccy) ?? 'USD';

		// Cross-tab live updates (the core API streams the same events over SSE in live mode).
		this.#channel = new BroadcastChannel('zv-plant-stream');
		this.#channel.onmessage = (e) => this.#apply(e.data, false);

		if (CORE_API) this.#connectLive();
		this.#tick = setInterval(() => this.#simulateTelemetry(), 2600);
	}

	destroy() {
		this.#channel?.close();
		if (this.#tick) clearInterval(this.#tick);
	}

	setTier(t: SubscriptionTier) {
		this.tier = t;
		if (browser) localStorage.setItem('zv-tier', t);
		if (t === 'importer_hub' && this.orgName === seed.ORG.name) this.orgName = seed.ORG.importer.name;
		if (t !== 'importer_hub' && this.orgName === seed.ORG.importer.name) this.orgName = seed.ORG.name;
		this.hints = {};
	}

	setCurrency(c: Ccy) {
		this.currency = c;
		localStorage.setItem('zv-ccy', c);
	}

	/* ---------------- silent hints ---------------- */

	hint(context: string, trigger: HintTrigger): SubtleHint | null {
		const h = evaluateSubtleHint(this.manifest, trigger, this.dismissals);
		this.hints[context] = h;
		return h;
	}
	clearHint(context: string) {
		this.hints[context] = null;
	}
	dismissHint(code: string) {
		this.dismissals[code] = Date.now();
		localStorage.setItem('zv-hint-dismissals', JSON.stringify(this.dismissals));
		for (const k of Object.keys(this.hints)) if (this.hints[k]?.code === code) this.hints[k] = null;
		if (this.mode === 'live') void this.#api('POST', '/api/v1/hints/dismiss', { hint_code: code });
	}

	/* ---------------- zones / telemetry ---------------- */

	updateZone(id: string, patch: Partial<seed.Zone>) {
		this.#apply({ type: 'zone', id, patch }, true);
		if (this.mode === 'live') void this.#api('PATCH', `/api/v1/zones/${id}/live-state`, patch);
	}

	logBatch(code: string, producedDelta: number, packagedDelta = 0) {
		this.#apply({ type: 'batch', code, producedDelta, packagedDelta }, true);
	}

	#apply(msg: { type: string; [k: string]: unknown }, broadcast: boolean) {
		if (msg.type === 'zone') {
			const z = this.zones.find((x) => x.id === msg.id);
			if (z) Object.assign(z, msg.patch);
		} else if (msg.type === 'batch') {
			const b = this.batches.find((x) => x.code === msg.code);
			if (b) {
				b.produced_qty = Math.min(b.target_qty, +(b.produced_qty + (msg.producedDelta as number)).toFixed(2));
				b.packaged_qty = Math.min(b.produced_qty, +(b.packaged_qty + (msg.packagedDelta as number)).toFixed(2));
				if (b.stage === 'scheduled' && b.produced_qty > 0) b.stage = 'in_production';
			}
		} else if (msg.type === 'notification') {
			const n = msg.notification as seed.Notification;
			if (!this.notifications.some((x) => x.id === n.id)) this.notifications.unshift(n);
		} else if (msg.type === 'read') {
			const n = this.notifications.find((x) => x.id === msg.id);
			if (n) n.read = true;
		}
		if (broadcast) this.#channel?.postMessage($state.snapshot(msg));
	}

	#simulateTelemetry() {
		for (const z of this.zones) {
			if (z.live_power_kw > 20) z.live_power_kw = Math.max(10, Math.round(z.live_power_kw + (Math.random() - 0.5) * z.live_power_kw * 0.04));
		}
		const live = this.batches.filter((b) => b.stage === 'in_production' && b.produced_qty < b.target_qty);
		for (const b of live) b.produced_qty = Math.min(b.target_qty, +(b.produced_qty + 0.01).toFixed(2));
		const pk = this.batches.find((b) => b.stage === 'packaging' && b.packaged_qty < b.produced_qty);
		if (pk) pk.packaged_qty = Math.min(pk.produced_qty, +(pk.packaged_qty + 0.01).toFixed(2));
	}

	/* ---------------- mail bell ---------------- */

	markRead(id: string) {
		this.#apply({ type: 'read', id }, true);
	}
	markAllRead(category?: string) {
		for (const n of this.notifications) if (!category || n.category === category) n.read = true;
	}
	snooze(id: string) {
		const n = this.notifications.find((x) => x.id === id);
		if (n?.deadline) n.deadline = new Date(new Date(n.deadline).getTime() + 86400000).toISOString();
		this.toast('Snoozed for 24 hours');
	}
	convertToRfq(id: string) {
		const n = this.notifications.find((x) => x.id === id);
		if (!n) return;
		const product = this.products.find((p) => n.tag?.includes(p.cn));
		const rfq = {
			id: `RFQ-${new Date().toISOString().slice(2, 7).replace('-', '')}-${String(this.connect.rfqs.length + 9).padStart(3, '0')}`,
			buyer: n.org_badge?.split(' · ')[0] ?? n.sender_email,
			product: product?.sku ?? 'ALB-6063',
			qty: Number(n.subject.match(/(\d+)\s*t\b/)?.[1] ?? 10),
			incoterm: 'FOB JNPT',
			status: 'open',
			quotes: [] as (typeof this.connect.rfqs)[number]['quotes']
		};
		this.connect.rfqs.unshift(rfq);
		n.read = true;
		n.link = { type: 'rfq', href: '/connect', label: `Open ${rfq.id}` };
		this.toast(`Converted to Zevra RFQ ${rfq.id}`, 'ok');
		if (this.mode === 'live') void this.#api('POST', `/api/v1/notifications/${id}/convert-rfq`, {});
	}

	/** Mock inbound webhook (inbox+<org>@mail.zevra.io) — classified like the FastAPI triage worker. */
	simulateInbound() {
		const samples = [
			{ name: 'Sofia Ricci', email: 's.ricci@milano-metalli.it', org: 'Milano Metalli · EU Importer', subject: 'Urgent RFQ: 25 t 6063 bars + CBAM XML by 10 Dec', snippet: 'We need a price quote for 25 t of 6063-T6 bars delivered CIF Genoa, plus your CBAM communication template and container specs.', tag: 'CN 76042910', deadline: new Date(Date.now() + 14 * 86400000).toISOString() },
			{ name: 'Jan de Vries', email: 'jan@rotterdam-sourcing.nl', org: 'Rotterdam Sourcing · Buying agent', subject: 'Sample order request — alloy sheets 5052', snippet: 'Our client in Eindhoven wants a sample order of 200 kg 5052-H32 sheets and the QR passport link.', tag: 'CN 76061220' }
		];
		const s = samples[Math.floor(Math.random() * samples.length)];
		const n: seed.Notification = {
			id: `n${Date.now()}`,
			category: 'buyer_agency',
			sender_name: s.name,
			sender_email: s.email,
			org_badge: s.org,
			subject: s.subject,
			snippet: s.snippet,
			tag: s.tag,
			deadline: s.deadline,
			severity: s.deadline ? 'elevated' : 'normal',
			link: { type: 'rfq', href: '/connect', label: 'Convert to Zevra RFQ' },
			read: false,
			received_at: new Date().toISOString()
		};
		this.#apply({ type: 'notification', notification: n }, true);
		this.toast(`New buyer email from ${s.org.split(' · ')[0]}`, 'ok');
	}

	/* ---------------- toasts ---------------- */

	toast(text: string, tone: Toast['tone'] = 'ok') {
		const id = ++this.#toastId;
		this.toasts.push({ id, text, tone });
		setTimeout(() => (this.toasts = this.toasts.filter((t) => t.id !== id)), 3800);
	}

	/* ---------------- live API ---------------- */

	async #api(method: string, path: string, body?: unknown) {
		try {
			const res = await fetch(`${CORE_API}${path}`, {
				method,
				credentials: 'include',
				headers: { 'content-type': 'application/json' },
				body: body ? JSON.stringify(body) : undefined
			});
			return res.ok ? await res.json().catch(() => ({})) : null;
		} catch {
			return null;
		}
	}

	async #connectLive() {
		const m = await this.#api('GET', '/api/v1/me/manifest');
		if (!m?.tier) return;
		this.mode = 'live';
		this.tier = m.tier;
		if (m.org_name) this.orgName = m.org_name;
		if (m.dismissed) this.dismissals = m.dismissed;
		const twin = await this.#api('GET', `/api/v1/plants/${m.plant_id ?? 'default'}/digital-twin`);
		if (twin?.zones?.length) this.zones = twin.zones;
		const notes = await this.#api('GET', '/api/v1/notifications');
		if (notes?.items) this.notifications = notes.items;
		try {
			const es = new EventSource(`${CORE_API}/api/v1/stream/plant/${m.plant_id ?? 'default'}`, { withCredentials: true });
			es.onmessage = (e) => this.#apply(JSON.parse(e.data), false);
		} catch {
			/* SSE optional */
		}
	}
}

export const ws = new Workspace();
