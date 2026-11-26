/**
 * ZENRES public-site content model.
 * All commercial terms and numeric claims mirror the ZENRES Complete Subscription & FreightLink
 * Blueprint (v2) and the Product Concept & Team Brief. Estimates are always labelled illustrative.
 */

export const NAV_LINKS = [
	{ label: 'Platform', href: '/#platform' },
	{ label: 'Features', href: '/features' },
	{ label: 'CBAM & Carbon', href: '/#cbam' },
	{ label: 'FreightLink', href: '/freightlink' },
	{ label: 'Traceability', href: '/#traceability' },
	{ label: 'Pricing', href: '/pricing' }
] as const;

/* ------------------------------------------------------------------ */
/* Disclaimers (verbatim wording used across the site)                 */
/* ------------------------------------------------------------------ */

export const DISCLAIMERS = {
	reporting:
		'ZENRES provides reporting support, not legal, tax or accounting advice. Final CBAM declarations remain the responsibility of the authorised declarant.',
	buyers:
		'Illustrative target buyer ecosystems (such as global retail and luxury groups) do not imply confirmed partnerships or endorsement.',
	anchoring:
		'Blockchain anchoring proves post-anchor tamper evidence, not source-input truthfulness.',
	cbamEstimate:
		'Illustrative estimate. Uses indicative default values, a sample measured intensity and an indicative certificate price. Not a tariff assessment.',
	freight:
		'Avoided-fuel figures are illustrative logistics estimates reported separately from statutory CBAM embedded emissions.',
	commercial:
		'Freight transport is paid separately. The optional 2% FreightLink commission is illustrative and always disclosed before booking. Basic carrier registration is free and independent of paid ranking.',
	pricing: 'All contract pricing is in USD, excluding applicable taxes. INR and EUR figures are indicative conversions.'
} as const;

/* ------------------------------------------------------------------ */
/* Pricing                                                            */
/* ------------------------------------------------------------------ */

export type Currency = 'USD' | 'INR' | 'EUR';
export type Billing = 'monthly' | 'annual';

/** Indicative display conversion only. Contracts are billed in USD. */
export const FX: Record<Currency, { rate: number; symbol: string; locale: string }> = {
	USD: { rate: 1, symbol: '$', locale: 'en-US' },
	INR: { rate: 88, symbol: '₹', locale: 'en-IN' },
	EUR: { rate: 0.86, symbol: '€', locale: 'de-DE' }
};

export function formatPrice(usd: number, currency: Currency): string {
	const fx = FX[currency];
	const value = Math.round(usd * fx.rate);
	return `${fx.symbol}${value.toLocaleString(fx.locale)}`;
}

export interface Tier {
	id: 'starter' | 'growth' | 'importer_hub';
	name: string;
	tagline: string;
	monthly: number;
	annual: number;
	flagship?: boolean;
	segment: string;
	coverage: string;
	ingestion: string;
	cbam: string;
	freight: string;
	addons: string;
	features: string[];
}

export const TIERS: Tier[] = [
	{
		id: 'starter',
		name: 'MSME Starter',
		tagline: 'Single-plant CBAM readiness',
		monthly: 99,
		annual: 990,
		segment: 'Single-plant MSMEs (~50–300t/yr to EU)',
		coverage: '1 plant • up to 5 CN-code SKUs',
		ingestion: 'Manual bills/PDFs/photos + Tally CSV',
		cbam: 'Basic Scope 1 & 2, Default vs Actual, XML/Excel, Email support',
		freight: 'Not included',
		addons: 'Optional $350 assisted setup sprint',
		features: [
			'1 plant, up to 5 CN-code SKUs',
			'Tally.ERP 9 CSV import',
			'Scope 1 & 2 activity capture',
			'Default vs Actual CBAM comparison',
			'EU Registry XML & Excel export',
			'Evidence vault & verifier marketplace',
			'Dual QR passports & ZENRES Connect profile'
		]
	},
	{
		id: 'growth',
		name: 'MSME Growth',
		tagline: 'Multi-line compliance + FreightLink',
		monthly: 249,
		annual: 2490,
		flagship: true,
		segment: 'Multi-line / multi-plant MSMEs (~300–1,500+t/yr)',
		coverage: 'Up to 3 plants • Unlimited CN-code SKUs',
		ingestion: 'OCR + WhatsApp + Vernacular floor-bot + Tally CSV',
		cbam: 'Full CN engine, Art. 9/CCTS, Precursor tracking, Tamper-evident Audit Locker, Capex ROI, Priority CA workspace',
		freight: 'Included: Empty truck finder, return-load matching, negotiation, QR pickup/POD, cost analytics',
		addons: '$350 setup fee per matrix; optional verifier/solar referrals',
		features: [
			'Up to 3 plants, unlimited CN-code SKUs',
			'OCR, WhatsApp & vernacular floor-bot ingestion',
			'Full CN engine with precursor tracking',
			'Article 9 / CCTS carbon-price deductions',
			'Tamper-evident hash-chained audit locker',
			'Decarbonisation Capex ROI simulator',
			'FreightLink backhaul matching & digital POD'
		]
	},
	{
		id: 'importer_hub',
		name: 'EU Importer Hub',
		tagline: 'B2B2B supplier network',
		monthly: 499,
		annual: 4990,
		segment: 'EU importers, brokers, trading houses & logistics operators',
		coverage: '15 suppliers included (15–500 portal; +$25/mo extra supplier)',
		ingestion: 'Bulk supplier invitation links, automated reminders, status tracker',
		cbam: 'Consolidated Registry XML pack, Portfolio carbon/tariff benchmarking, Dedicated AM & EU Customs Audit Defence Pack',
		freight: 'Advanced: Multi-truck live fleet, AI/bulk backhaul, cargo consolidation, international coordination, green logistics score',
		addons: 'Extra suppliers $25/mo; custom ERP/fleet integrations quoted',
		features: [
			'15 suppliers included, scales to 500',
			'Bulk invitation links & reminder cadences',
			'Consolidated multi-supplier Registry XML',
			'Supplier carbon & tariff benchmarking',
			'EU Customs Audit Defence Pack',
			'Live multi-truck fleet & cargo consolidation',
			'Dedicated account manager'
		]
	}
];

export const COMPARISON: { dimension: string; values: [string, string, string] }[] = [
	{ dimension: 'Monthly / Annual', values: ['$99 / mo • $990 / yr', '$249 / mo • $2,490 / yr', '$499 / mo • $4,990 / yr'] },
	{ dimension: 'Target segment', values: ['Single-plant MSMEs (~50–300t/yr to EU)', 'Multi-line / multi-plant MSMEs (~300–1,500+t/yr)', 'EU importers, brokers, trading houses & logistics operators'] },
	{ dimension: 'Plants', values: ['1', 'Up to 3', 'Supplier network'] },
	{ dimension: 'CN-code SKUs', values: ['Up to 5', 'Unlimited', 'Unlimited across suppliers'] },
	{ dimension: 'Supplier portal', values: ['—', '—', '15 included • 15–500 • +$25/mo per extra supplier'] },
	{ dimension: 'Data ingestion', values: ['Manual bills/PDFs/photos + Tally CSV', 'OCR + WhatsApp + Vernacular floor-bot + Tally CSV', 'Bulk supplier invitation links, automated reminders, status tracker'] },
	{ dimension: 'Tally.ERP 9 integration', values: ['CSV / Excel import', 'CSV + XML / ODBC middleware + freight sync', 'Per-supplier sync + custom ERP quoted'] },
	{ dimension: 'Scope 1 & 2 • Default vs Actual', values: ['Basic', 'Full', 'Consolidated portfolio'] },
	{ dimension: 'Full CN engine • Art. 9 / CCTS • Precursors', values: ['—', 'Included', 'Included'] },
	{ dimension: 'Audit Evidence Locker', values: ['Central vault + verifier marketplace', 'Tamper-evident hash trail + Priority CA workspace', 'EU Customs Audit Defence Pack + dedicated AM'] },
	{ dimension: 'Decarbonisation Capex ROI simulator', values: ['—', 'Included', 'Included'] },
	{ dimension: 'Dual QR traceability • ZENRES Connect', values: ['Included', 'Included', 'Included'] },
	{ dimension: 'FreightLink logistics', values: ['Not included', 'Empty truck finder, return-load matching, negotiation, QR pickup/POD, cost analytics', 'Multi-truck live fleet, AI/bulk backhaul, cargo consolidation, international coordination, green logistics score'] },
	{ dimension: 'Support & setup', values: ['Email support • optional $350 assisted setup sprint', 'Priority CA workspace • $350 setup fee per matrix', 'Dedicated account manager • custom integrations quoted'] }
];

/* ------------------------------------------------------------------ */
/* Four pillars                                                       */
/* ------------------------------------------------------------------ */

export const PILLARS = [
	{
		index: '01',
		name: 'ZENRES Business',
		headline: "Connect Tally, don't replace it.",
		body: 'Your books stay in Tally.ERP 9. ZENRES reads sales and purchase ledgers, BOMs and inventory movement so production costing, stock and the 3D factory view stay in step with the records you already trust.',
		capabilities: ['Tally.ERP 9 CSV / XML / ODBC sync', 'Sales & purchase ledgers', 'Bills of materials', 'Inventory movement', 'Production costing', '3D factory operational visibility'],
		metric: { label: 'Voucher dedupe', value: 'Idempotent' }
	},
	{
		index: '02',
		name: 'ZENRES Carbon',
		headline: 'Audit-ready EU CBAM reporting.',
		body: 'Capture Scope 1 & 2 activity from bills, meters and the factory floor. Map every product to its CN code, apply date-effective rules and export Registry-ready XML — with every figure labelled for data quality.',
		capabilities: ['Scope 1 & 2 activity capture', 'OCR & WhatsApp floor-bot ingestion', 'CN-code engine', 'Article 9 / CCTS deductions', 'Registry XML / Excel export', 'Measured vs estimated labels'],
		metric: { label: 'Data-quality tiers', value: '4' }
	},
	{
		index: '03',
		name: 'ZENRES Trace & FreightLink',
		headline: 'Dual QR passports & zero empty return km.',
		body: 'Product/Batch QR and Container/Shipment QR are linked in one ledger, with selective off-chain hash anchoring. FreightLink fills the empty return leg of every corridor with matched cargo.',
		capabilities: ['Product / Batch QR', 'Container / Shipment QR', 'Selective off-chain hash anchoring', 'Corridor backhaul matching', 'QR pickup & delivery', 'Digital proof of delivery'],
		metric: { label: 'Linked QR systems', value: '2' }
	},
	{
		index: '04',
		name: 'ZENRES Connect',
		headline: 'Discoverable to global buyers.',
		body: 'Verified MSME profiles and carbon-enriched catalogues put export-ready factories in front of international buyers — with structured RFQs and an EU Importer portal that scales from 15 to 500 suppliers.',
		capabilities: ['Verified MSME profiles', 'Carbon-enriched catalogues', 'Structured RFQ workflows', 'Side-by-side quote comparison', 'EU Importer multi-supplier portal', 'Disclosed intermediary introductions'],
		metric: { label: 'Supplier portal', value: '15–500' }
	}
] as const;

/* ------------------------------------------------------------------ */
/* Eight-step journey                                                 */
/* ------------------------------------------------------------------ */

export const JOURNEY = [
	{ n: '01', title: 'Business records in Tally', body: 'Sales, purchases, stock and BOMs sync from Tally.ERP 9 via CSV, XML or ODBC — never replaced.', tag: 'ZENRES Business' },
	{ n: '02', title: 'Production & utility capture', body: 'Electricity, fuel and process inputs from bills, meters, OCR and WhatsApp floor supervisors.', tag: 'ZENRES Carbon' },
	{ n: '03', title: 'Carbon calculation', body: 'Direct and indirect embedded emissions per CN code, with date-effective rules and quality labels.', tag: 'ZENRES Carbon' },
	{ n: '04', title: 'Product / Batch QR', body: 'Each batch receives a non-guessable passport linking specs, production date and carbon footprint.', tag: 'ZENRES Trace' },
	{ n: '05', title: 'Container / Shipment QR', body: 'Containers such as ZV-001 link the batches packed inside, with dispatch status and documents.', tag: 'ZENRES Trace' },
	{ n: '06', title: 'Verifiable traceability', body: 'Selective off-chain hash anchoring makes post-anchor edits tamper-evident.', tag: 'ZENRES Trace' },
	{ n: '07', title: 'Buyer discovery', body: 'Approved evidence enriches your ZENRES Connect catalogue for international buyers and RFQs.', tag: 'ZENRES Connect' },
	{ n: '08', title: 'Export & CBAM submission support', body: 'Registry-ready XML and Excel packs for your EU importer, with an audit trail behind every number.', tag: 'ZENRES Carbon' }
] as const;

/* ------------------------------------------------------------------ */
/* CBAM illustrative model                                            */
/* ------------------------------------------------------------------ */

export const CBAM_MODEL = {
	product: 'CN 7604 — Aluminium bars & profiles',
	/** Indicative default value incl. date-effective markup (tCO2e per tonne). */
	defaultIntensity: 7.9,
	/** Sample measured actual embedded emissions (tCO2e per tonne). */
	actualIntensity: 5.4,
	/** Indicative CBAM certificate price (EUR per tCO2e). */
	certificatePrice: 76,
	minTonnes: 50,
	maxTonnes: 1500
};

export function cbamExposure(tonnes: number) {
	const def = tonnes * CBAM_MODEL.defaultIntensity * CBAM_MODEL.certificatePrice;
	const act = tonnes * CBAM_MODEL.actualIntensity * CBAM_MODEL.certificatePrice;
	return {
		defaultEur: def,
		actualEur: act,
		deltaEur: def - act,
		defaultT: tonnes * CBAM_MODEL.defaultIntensity,
		actualT: tonnes * CBAM_MODEL.actualIntensity
	};
}

/* ------------------------------------------------------------------ */
/* FreightLink corridors                                              */
/* ------------------------------------------------------------------ */

export const DIESEL_KG_CO2_PER_L = 2.68;

export interface Corridor {
	id: 'DEL-BOM' | 'LDH-DEL';
	label: string;
	from: { city: string; code: string; x: number; y: number };
	to: { city: string; code: string; x: number; y: number };
	km: number;
	outbound: string;
	returnCargo: string;
	vehicle: string;
	payload: string;
	baselineFuelL: number;
	matchedFuelL: number;
}

export const CORRIDORS: Corridor[] = [
	{
		id: 'DEL-BOM',
		label: 'Delhi ↔ Mumbai',
		from: { city: 'Delhi', code: 'DEL', x: 300, y: 110 },
		to: { city: 'Mumbai', code: 'BOM', x: 150, y: 330 },
		km: 1415,
		outbound: 'Aluminium profiles, 18 t — Deccan Alloy Works',
		returnCargo: 'Packaged machine parts, 14 t — Bhiwandi',
		vehicle: '32 ft multi-axle container',
		payload: '21 t rated',
		baselineFuelL: 180,
		matchedFuelL: 140
	},
	{
		id: 'LDH-DEL',
		label: 'Ludhiana ↔ Delhi',
		from: { city: 'Ludhiana', code: 'LDH', x: 250, y: 40 },
		to: { city: 'Delhi', code: 'DEL', x: 300, y: 110 },
		km: 310,
		outbound: 'Steel fasteners, 9 t — Focal Point',
		returnCargo: 'Textile cartons, 7 t — Okhla',
		vehicle: '19 ft closed body',
		payload: '10 t rated',
		baselineFuelL: 64,
		matchedFuelL: 51
	}
];

export const FREIGHT_STAGES = [
	{ key: 'outbound', name: 'Outbound', detail: 'Loaded truck departs origin with the shipper’s export consignment.', filter: 'Shipment booked • consignment QR attached' },
	{ key: 'delivery', name: 'Delivery', detail: 'Consignment delivered and scanned at destination. Truck is now empty.', filter: 'Delivery QR scan • POD captured' },
	{ key: 'capacity', name: 'Return capacity listing', detail: 'Carrier lists the empty return leg with window, payload and vehicle type. Stale listings auto-expire.', filter: 'Corridor • departure window • payload • vehicle type' },
	{ key: 'demand', name: 'Demand post', detail: 'A shipper on the return corridor posts load demand: pickup, drop, weight, cargo and dates.', filter: 'Pickup / drop • weight • dimensions • cargo type' },
	{ key: 'match', name: 'Corridor match', detail: 'Hard filters first — corridor, window, payload, vehicle, cargo rules, valid documents — then ranked by detour km, availability, quote and reliability.', filter: 'Hard filters → rank(detour, availability, quote, reliability)' },
	{ key: 'booking', name: 'Price negotiation & booking', detail: 'Itemised quote with taxes, extras and cancellation terms. Booking requires explicit authorised acceptance — never auto-booked.', filter: 'Itemised quote • disclosed fee • explicit acceptance' },
	{ key: 'transit', name: 'QR pickup & transit', detail: 'Pickup QR scanned at loading. Consent-based GPS tracking only when the carrier opts in.', filter: 'Pickup QR • GPS consent • milestone events' },
	{ key: 'pod', name: 'Digital POD & settlement', detail: 'Digital proof of delivery closes the trip; freight invoice syncs to Tally for reconciliation.', filter: 'Digital POD • freight invoice → Tally' }
] as const;

/* ------------------------------------------------------------------ */
/* QR passport demo                                                   */
/* ------------------------------------------------------------------ */

export const PASSPORT_DEMO = {
	container: { code: 'ZV-001', token: '01JD8ZV001CNTR7Q4M2X9KPWBA', status: 'Ready to dispatch', route: 'Pune → JNPT Nhava Sheva → Rotterdam' },
	batches: [
		{ code: 'A-101', token: '01JD8ZA101BTCH3N6R1VQ8MZTE', product: 'Aluminium bars', cn: '76042910', qty: '6.0 t', tco2e: '5.38', quality: 'Measured' },
		{ code: 'A-102', token: '01JD8ZA102BTCH9H2K7WD4FXRC', product: 'Aluminium bars', cn: '76042910', qty: '5.5 t', tco2e: '5.41', quality: 'Measured' },
		{ code: 'A-103', token: '01JD8ZA103BTCH5T8P3LJ6NYGS', product: 'Extruded profiles', cn: '76042100', qty: '6.5 t', tco2e: '5.62', quality: 'Supplier-reported' }
	],
	views: [
		{ id: 'public', name: 'Public buyer view', fields: ['Manufacturer', 'Product & CN code', 'Batch & production date', 'Approved carbon badge', 'Verification tier'] },
		{ id: 'partner', name: 'Authenticated partner view', fields: ['Everything public', 'Linked batches & quantities', 'Dispatch milestones', 'Transport documents', 'Calculation summary'] },
		{ id: 'private', name: 'Private manufacturer view', fields: ['Everything partner', 'BOM & precursor lots', 'Meter evidence', 'Costing', 'Audit hash trail'] }
	]
};
