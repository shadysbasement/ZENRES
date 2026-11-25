/**
 * Deterministic FreightLink corridor matcher (mirrors services/core-api/internal/freight).
 * Hard filters first — corridor, window overlap, payload, vehicle/cargo rules, valid docs —
 * then rank by detour km, availability, quote and reliability. Paid promotion is NOT an input.
 */
export interface Truck {
	id: string;
	reg: string;
	carrier: string;
	vehicle: string;
	payload: number;
	corridor: string;
	window: [string, string];
	detour_km: number;
	quote_inr: number;
	reliability: number;
	verified: boolean;
	docs_valid: boolean;
	promoted?: boolean;
}
export interface Load {
	id: string;
	corridor: string;
	weight: number;
	cargo: string;
	window: [string, string];
	budget_inr: number;
	requires_closed?: boolean;
}
export interface Match {
	truck: Truck;
	score: number;
	reasons: string[];
	rejected?: string;
}

const overlaps = (a: [string, string], b: [string, string]) => new Date(a[0]) <= new Date(b[1]) && new Date(b[0]) <= new Date(a[1]);
const CLOSED = /(container|closed)/i;

export function hardFilter(t: Truck, l: Load): string | null {
	if (t.corridor !== l.corridor) return 'corridor';
	if (!overlaps(t.window, l.window)) return 'window';
	if (t.payload < l.weight) return 'payload';
	if (l.requires_closed && !CLOSED.test(t.vehicle)) return 'vehicle_type';
	if (!t.docs_valid) return 'documents';
	return null;
}

export function rankMatches(trucks: Truck[], load: Load): Match[] {
	const ok: Match[] = [];
	for (const t of trucks) {
		const rej = hardFilter(t, load);
		if (rej) continue;
		// Normalised components (0..1), weights sum to 1. `promoted` deliberately ignored.
		const detour = 1 - Math.min(1, t.detour_km / 100);
		const quote = Math.max(0, Math.min(1, 1 - (t.quote_inr - load.budget_inr * 0.7) / (load.budget_inr * 0.6)));
		const reliability = Math.max(0, (t.reliability - 3) / 2);
		const availability = 1;
		const score = +(detour * 0.35 + quote * 0.3 + reliability * 0.25 + availability * 0.1).toFixed(4);
		ok.push({ truck: t, score, reasons: [`${t.detour_km} km detour`, `₹${t.quote_inr.toLocaleString('en-IN')} quote`, `★ ${t.reliability}`] });
	}
	return ok.sort((a, b) => b.score - a.score || a.truck.detour_km - b.truck.detour_km);
}

export function rejected(trucks: Truck[], load: Load) {
	return trucks.map((t) => ({ truck: t, reason: hardFilter(t, load) })).filter((x) => x.reason);
}

/** A booking can only be created from an explicit, authorised acceptance — never automatically. */
export function createBooking(match: Match, acceptance: { accepted: boolean; authorisedBy?: string; feePct: number }) {
	if (!acceptance.accepted || !acceptance.authorisedBy) throw new Error('explicit_acceptance_required');
	const fee = +(match.truck.quote_inr * (acceptance.feePct / 100)).toFixed(0);
	return { status: 'confirmed' as const, freight_inr: match.truck.quote_inr, disclosed_fee_inr: fee, total_inr: match.truck.quote_inr + fee, authorised_by: acceptance.authorisedBy };
}
