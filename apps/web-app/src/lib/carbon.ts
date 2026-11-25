/**
 * Deterministic CBAM / carbon math. Mirrors services/core-api/internal/cbam so the browser
 * preview and the server-side calculation workbench always agree.
 */

/** Indicative CBAM certificate price (EUR / tCO2e) used for illustrative exposure figures. */
export const CERT_PRICE_EUR = 76;
export const DIESEL_KG_CO2_PER_L = 2.68;
export const USD_INR = 88;
export const USD_EUR = 0.86;

export type ProductEmissions = { direct_tco2e_t: number; indirect_tco2e_t: number; precursor_tco2e_t: number; default_tco2e_t: number };

/** Specific embedded emissions (tCO2e per tonne of product) from measured/supplier data. */
export function specificEmbedded(p: ProductEmissions, opts: { includePrecursors?: boolean } = {}): number {
	const pre = opts.includePrecursors === false ? 0 : p.precursor_tco2e_t;
	return round(p.direct_tco2e_t + p.indirect_tco2e_t + pre, 4);
}

/** Gross CBAM exposure before free-allocation adjustment (EUR). */
export function exposureEur(tonnes: number, intensity: number, price = CERT_PRICE_EUR): number {
	return round(tonnes * intensity * price, 2);
}

/**
 * Article 9 deduction for carbon price effectively paid in the country of origin (e.g. Indian CCTS),
 * capped so the deduction never exceeds the CBAM certificate obligation.
 */
export function article9Deduction(embeddedT: number, pricePaidEurPerT: number, price = CERT_PRICE_EUR): number {
	const capped = Math.min(Math.max(0, pricePaidEurPerT), price);
	return round(embeddedT * capped, 2);
}

export function compareDefaultVsActual(tonnes: number, p: ProductEmissions, price = CERT_PRICE_EUR) {
	const actual = specificEmbedded(p);
	const def = p.default_tco2e_t;
	const defaultEur = exposureEur(tonnes, def, price);
	const actualEur = exposureEur(tonnes, actual, price);
	return { actual, def, defaultEur, actualEur, savingsEur: round(defaultEur - actualEur, 2), reductionPct: round(((def - actual) / def) * 100, 1) };
}

/** Activity record → tCO2e (emission factor already expressed in tCO2e per activity unit). */
export function activityEmissions(qty: number, efTPerUnit: number): number {
	return round(qty * efTPerUnit, 4);
}

/**
 * FreightLink avoided-fuel estimate. Reported SEPARATELY from CBAM embedded emissions and
 * never netted against them (guards against double counting).
 */
export function avoidedFuel(baselineL: number, matchedL: number) {
	const litres = Math.max(0, baselineL - matchedL);
	return { litres, kgCO2: round(litres * DIESEL_KG_CO2_PER_L, 1), scope: 'logistics_estimate' as const };
}

/** Capex ROI simulator: simple payback, NPV-based IRR (bisection) and avoided tonnes. */
export function capexScenario(input: { capexUsd: number; annualSavingsUsd: number; annualAvoidedT: number; years: number; depreciationPct: number; taxRatePct: number; certPriceEur?: number; exportShare?: number }) {
	const price = input.certPriceEur ?? CERT_PRICE_EUR;
	const cbamSavingsUsd = (input.annualAvoidedT * (input.exportShare ?? 0.7) * price) / USD_EUR;
	const taxShield = (input.capexUsd * (input.depreciationPct / 100) * (input.taxRatePct / 100));
	const yearly = input.annualSavingsUsd + cbamSavingsUsd;
	const flows = [-input.capexUsd, ...Array.from({ length: input.years }, (_, i) => yearly + (i === 0 ? taxShield : 0))];
	let cum = -input.capexUsd;
	let payback: number | null = null;
	const cumulative = [cum];
	for (let y = 1; y <= input.years; y++) {
		const prev = cum;
		cum += flows[y];
		cumulative.push(round(cum, 0));
		if (payback === null && cum >= 0) payback = round(y - 1 + -prev / flows[y], 2);
	}
	return { payback, irr: irr(flows), cbamSavingsUsd: round(cbamSavingsUsd, 0), yearly: round(yearly, 0), cumulative, avoidedT: round(input.annualAvoidedT * input.years, 1) };
}

export function irr(flows: number[]): number | null {
	const npv = (r: number) => flows.reduce((s, f, i) => s + f / Math.pow(1 + r, i), 0);
	let lo = -0.99,
		hi = 5;
	if (npv(lo) * npv(hi) > 0) return null;
	for (let i = 0; i < 120; i++) {
		const mid = (lo + hi) / 2;
		if (npv(lo) * npv(mid) <= 0) hi = mid;
		else lo = mid;
	}
	return round(((lo + hi) / 2) * 100, 1);
}

/* ---------------- Tamper-evident hash chain ---------------- */

export async function sha256Hex(text: string): Promise<string> {
	const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));
	return [...new Uint8Array(buf)].map((b) => b.toString(16).padStart(2, '0')).join('');
}

/** Canonical JSON: sorted keys, no whitespace — identical to the Go implementation. */
export function canonical(value: unknown): string {
	if (value === null || typeof value !== 'object') return JSON.stringify(value);
	if (Array.isArray(value)) return `[${value.map(canonical).join(',')}]`;
	const obj = value as Record<string, unknown>;
	return `{${Object.keys(obj)
		.sort()
		.map((k) => `${JSON.stringify(k)}:${canonical(obj[k])}`)
		.join(',')}}`;
}

export const GENESIS = '0'.repeat(64);

export async function buildChain<T extends object>(records: T[]): Promise<(T & { prev_hash: string; hash: string })[]> {
	let prev = GENESIS;
	const out: (T & { prev_hash: string; hash: string })[] = [];
	for (const r of records) {
		const hash = await sha256Hex(prev + canonical(r));
		out.push({ ...r, prev_hash: prev, hash });
		prev = hash;
	}
	return out;
}

/** Returns the index of the first broken link, or -1 when the chain is intact. */
export async function verifyChain<T extends { prev_hash: string; hash: string }>(chain: T[]): Promise<number> {
	let prev = GENESIS;
	for (let i = 0; i < chain.length; i++) {
		const { prev_hash, hash, ...rest } = chain[i];
		if (prev_hash !== prev) return i;
		if ((await sha256Hex(prev + canonical(rest))) !== hash) return i;
		prev = hash;
	}
	return -1;
}

export function round(v: number, d = 2): number {
	const f = Math.pow(10, d);
	return Math.round(v * f) / f;
}
