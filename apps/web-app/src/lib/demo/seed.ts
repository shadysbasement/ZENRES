/**
 * Pilot seed — "Deccan Alloy Works" Indian aluminium MSME.
 * Mirrors services/core-api/cmd/seed-pilot so demo mode and a seeded Postgres show identical data.
 */

export type ZoneType = 'production' | 'packaging' | 'office' | 'warehouse';
export type DataQuality = 'measured' | 'supplier_reported' | 'estimated' | 'verified';

export interface Zone {
	id: string;
	code: string;
	name: string;
	type: ZoneType;
	pos: [number, number, number];
	size: [number, number, number];
	supervisor: string;
	active_workers: number;
	shift_target_workers: number;
	live_power_kw: number;
	fuel_lph: number;
	staff?: Record<string, number>;
	notes?: string[];
}

export interface Product {
	id: string;
	sku: string;
	cn: string;
	name: string;
	sector: string;
	unit_weight_kg: number;
	capacity_tpm: number;
	unit_cost_usd: number;
	wastage_pct: number;
	cost: { material: number; labour: number; energy: number; overhead: number };
	bom: { input: string; qty_per_t: number; unit: string; precursor: boolean; supplier?: string; tco2e_per_t?: number }[];
	published_connect: boolean;
	direct_tco2e_t: number;
	indirect_tco2e_t: number;
	precursor_tco2e_t: number;
	default_tco2e_t: number;
	quality: DataQuality;
}

export interface Batch {
	id: string;
	code: string;
	product_id: string;
	zone_code: string;
	stage: 'scheduled' | 'in_production' | 'packaging' | 'ready_dispatch' | 'shipped';
	target_qty: number;
	produced_qty: number;
	packaged_qty: number;
	direct_tco2e: number;
	indirect_tco2e: number;
	quality: DataQuality;
	produced_on: string;
	qr_token: string;
}

export interface InventoryItem {
	id: string;
	code: string;
	name: string;
	category: 'raw_material' | 'wip' | 'finished_good' | 'consumable';
	qty: number;
	unit: string;
	reorder: number;
	unit_value_usd: number;
	tally_id: string;
	lot?: string;
}

export interface Voucher {
	id: string;
	source: 'manual' | 'tally_csv' | 'tally_xml' | 'tally_odbc' | 'freightlink';
	ref: string;
	type: 'sales' | 'purchase' | 'expense' | 'utility' | 'freight';
	counterparty: string;
	date: string;
	amount_usd: number;
	cost_usd: number;
	is_export: boolean;
}

export interface Notification {
	id: string;
	category: 'buyer_agency' | 'deadline' | 'operational_alert';
	sender_name: string;
	sender_email: string;
	org_badge?: string;
	subject: string;
	snippet: string;
	body?: string;
	tag?: string;
	deadline?: string;
	severity: 'normal' | 'elevated' | 'critical';
	link?: { type: string; href: string; label: string };
	read: boolean;
	received_at: string;
}

const iso = (d: Date) => d.toISOString();
const NOW = new Date('2026-11-26T09:30:00+05:30');
const ago = (mins: number) => iso(new Date(NOW.getTime() - mins * 60_000));
const ahead = (days: number) => iso(new Date(NOW.getTime() + days * 86_400_000));

export const ORG = {
	id: 'org_deccan',
	name: 'Deccan Alloy Works Pvt. Ltd.',
	short: 'Deccan Alloy Works',
	gstin: '27AAACD4821K1Z7',
	city: 'Chakan MIDC, Pune',
	inbound: 'inbox+deccan-alloy@mail.zenres.example',
	importer: { name: 'Rhein Aluminium Import GmbH', short: 'Rhein Aluminium Import', eori: 'DE472918350000' }
};

export const USER = { name: 'Ananya Deshpande', email: 'ananya@deccanalloy.in', role: 'owner', initials: 'AD' };

export const PLANTS = [
	{
		id: 'plt_chakan',
		name: 'Chakan Extrusion Plant',
		city: 'Chakan MIDC, Pune',
		state: 'Maharashtra',
		un_locode: 'INPNQ',
		grid_region: 'IN-WEST',
		meters: ['MSEDCL-HT-4471', 'DG-SET-02'],
		dims: { width: 60, depth: 40, height: 12 }
	}
];

export const ZONES: Zone[] = [
	{ id: 'z_proda', code: 'PROD-A', name: 'Line A · Smelting & Casting', type: 'production', pos: [-17, 0, -8], size: [24, 9, 16], supervisor: 'Rakesh Patil', active_workers: 11, shift_target_workers: 12, live_power_kw: 612, fuel_lph: 38 },
	{ id: 'z_prodb', code: 'PROD-B', name: 'Line B · Machining & Finishing', type: 'production', pos: [9, 0, -8], size: [20, 7, 16], supervisor: 'Sunita Kale', active_workers: 7, shift_target_workers: 8, live_power_kw: 284, fuel_lph: 0 },
	{ id: 'z_pack', code: 'PACK-1', name: 'Packaging & Dispatch Bay', type: 'packaging', pos: [9, 0, 11], size: [20, 6, 14], supervisor: 'Imran Shaikh', active_workers: 9, shift_target_workers: 10, live_power_kw: 46, fuel_lph: 0 },
	{ id: 'z_off', code: 'OFF-ADMIN', name: 'Offices & Admin Wing', type: 'office', pos: [-23, 0, 11], size: [12, 8, 14], supervisor: 'Meera Joshi', active_workers: 6, shift_target_workers: 6, live_power_kw: 18, fuel_lph: 0, staff: { Accounts: 2, Compliance: 1, QA: 2, Management: 1 } },
	{ id: 'z_wh', code: 'WH-MAIN', name: 'Raw Material & FG Warehouse', type: 'warehouse', pos: [-8, 0, 11], size: [12, 8, 14], supervisor: 'Vikram Rao', active_workers: 3, shift_target_workers: 4, live_power_kw: 9, fuel_lph: 2 }
];

export const PRODUCTS: Product[] = [
	{
		id: 'prd_alb6063',
		sku: 'ALB-6063',
		cn: '76042910',
		name: 'Aluminium bars 6063-T6',
		sector: 'Aluminium',
		unit_weight_kg: 1000,
		capacity_tpm: 140,
		unit_cost_usd: 2860,
		wastage_pct: 3.2,
		cost: { material: 2140, labour: 180, energy: 390, overhead: 150 },
		bom: [
			{ input: 'Primary aluminium ingot P1020', qty_per_t: 0.78, unit: 't', precursor: true, supplier: 'Hindalco (Renukoot)', tco2e_per_t: 3.9 },
			{ input: 'Post-consumer aluminium scrap', qty_per_t: 0.25, unit: 't', precursor: true, supplier: 'Local scrap aggregators', tco2e_per_t: 0.5 },
			{ input: 'Magnesium', qty_per_t: 0.006, unit: 't', precursor: false },
			{ input: 'Silicon metal', qty_per_t: 0.005, unit: 't', precursor: false }
		],
		published_connect: true,
		direct_tco2e_t: 0.62,
		indirect_tco2e_t: 1.71,
		precursor_tco2e_t: 3.17,
		default_tco2e_t: 7.9,
		quality: 'measured'
	},
	{
		id: 'prd_alp6061',
		sku: 'ALP-6061',
		cn: '76042100',
		name: 'Extruded hollow profiles 6061',
		sector: 'Aluminium',
		unit_weight_kg: 1000,
		capacity_tpm: 90,
		unit_cost_usd: 3120,
		wastage_pct: 4.1,
		cost: { material: 2290, labour: 230, energy: 430, overhead: 170 },
		bom: [
			{ input: 'Primary aluminium ingot P1020', qty_per_t: 0.85, unit: 't', precursor: true, supplier: 'Hindalco (Renukoot)', tco2e_per_t: 3.9 },
			{ input: 'Post-consumer aluminium scrap', qty_per_t: 0.19, unit: 't', precursor: true, supplier: 'Local scrap aggregators', tco2e_per_t: 0.5 }
		],
		published_connect: true,
		direct_tco2e_t: 0.66,
		indirect_tco2e_t: 1.82,
		precursor_tco2e_t: 3.41,
		default_tco2e_t: 7.9,
		quality: 'supplier_reported'
	},
	{
		id: 'prd_alw1350',
		sku: 'ALW-1350',
		cn: '76051100',
		name: 'Aluminium wire EC grade 1350',
		sector: 'Aluminium',
		unit_weight_kg: 1000,
		capacity_tpm: 60,
		unit_cost_usd: 2780,
		wastage_pct: 2.4,
		cost: { material: 2120, labour: 160, energy: 360, overhead: 140 },
		bom: [{ input: 'Primary aluminium ingot P1020', qty_per_t: 1.02, unit: 't', precursor: true, supplier: 'NALCO (Angul)', tco2e_per_t: 4.2 }],
		published_connect: false,
		direct_tco2e_t: 0.41,
		indirect_tco2e_t: 1.55,
		precursor_tco2e_t: 4.28,
		default_tco2e_t: 8.4,
		quality: 'estimated'
	},
	{
		id: 'prd_als5052',
		sku: 'ALS-5052',
		cn: '76061220',
		name: 'Alloy sheets 5052-H32',
		sector: 'Aluminium',
		unit_weight_kg: 1000,
		capacity_tpm: 45,
		unit_cost_usd: 3340,
		wastage_pct: 5.0,
		cost: { material: 2400, labour: 260, energy: 480, overhead: 200 },
		bom: [
			{ input: 'Primary aluminium ingot P1020', qty_per_t: 0.9, unit: 't', precursor: true, supplier: 'Hindalco (Renukoot)', tco2e_per_t: 3.9 },
			{ input: 'Magnesium', qty_per_t: 0.025, unit: 't', precursor: false }
		],
		published_connect: true,
		direct_tco2e_t: 0.71,
		indirect_tco2e_t: 2.04,
		precursor_tco2e_t: 3.51,
		default_tco2e_t: 8.1,
		quality: 'measured'
	}
];

export const BATCHES: Batch[] = [
	{ id: 'bat_a101', code: 'A-101', product_id: 'prd_alb6063', zone_code: 'PACK-1', stage: 'ready_dispatch', target_qty: 6, produced_qty: 6, packaged_qty: 6, direct_tco2e: 3.72, indirect_tco2e: 10.26, quality: 'measured', produced_on: '2026-11-09', qr_token: '01JD8ZA101BTCH3N6R1VQ8MZTE' },
	{ id: 'bat_a102', code: 'A-102', product_id: 'prd_alb6063', zone_code: 'PACK-1', stage: 'ready_dispatch', target_qty: 5.5, produced_qty: 5.5, packaged_qty: 5.5, direct_tco2e: 3.41, indirect_tco2e: 9.41, quality: 'measured', produced_on: '2026-11-11', qr_token: '01JD8ZA102BTCH9H2K7WD4FXRC' },
	{ id: 'bat_a103', code: 'A-103', product_id: 'prd_alp6061', zone_code: 'PACK-1', stage: 'ready_dispatch', target_qty: 6.5, produced_qty: 6.5, packaged_qty: 6.5, direct_tco2e: 4.29, indirect_tco2e: 11.83, quality: 'supplier_reported', produced_on: '2026-11-13', qr_token: '01JD8ZA103BTCH5T8P3LJ6NYGS' },
	{ id: 'bat_a104', code: 'A-104', product_id: 'prd_alb6063', zone_code: 'PROD-A', stage: 'in_production', target_qty: 8, produced_qty: 4.2, packaged_qty: 0, direct_tco2e: 2.6, indirect_tco2e: 7.18, quality: 'measured', produced_on: '2026-11-25', qr_token: '01JD8ZA104BTCH7Y2M5QK8RWDN' },
	{ id: 'bat_a105', code: 'A-105', product_id: 'prd_als5052', zone_code: 'PROD-B', stage: 'in_production', target_qty: 4, produced_qty: 1.6, packaged_qty: 0, direct_tco2e: 1.14, indirect_tco2e: 3.26, quality: 'measured', produced_on: '2026-11-25', qr_token: '01JD8ZA105BTCH2C9V6TB3LXQH' },
	{ id: 'bat_a106', code: 'A-106', product_id: 'prd_alw1350', zone_code: 'PACK-1', stage: 'packaging', target_qty: 3, produced_qty: 3, packaged_qty: 1.9, direct_tco2e: 1.23, indirect_tco2e: 4.65, quality: 'estimated', produced_on: '2026-11-22', qr_token: '01JD8ZA106BTCH8F4N1HP7MZKA' },
	{ id: 'bat_a107', code: 'A-107', product_id: 'prd_alp6061', zone_code: 'PROD-A', stage: 'scheduled', target_qty: 7, produced_qty: 0, packaged_qty: 0, direct_tco2e: 0, indirect_tco2e: 0, quality: 'estimated', produced_on: '2026-11-28', qr_token: '01JD8ZA107BTCH4K8W2DX6NQTR' }
];

export const CONTAINERS = [
	{ id: 'cnt_zv001', code: 'ZV-001', token: '01JD8ZV001CNTR7Q4M2X9KPWBA', status: 'active' as 'active' | 'reprinted' | 'revoked', dispatch: 'Ready to dispatch', batches: ['A-101', 'A-102', 'A-103'], route: 'Pune → JNPT Nhava Sheva → Rotterdam', anchored_at: '2026-11-18T09:42:00Z' as string | null, merkle_root: '9f2c4be1d07a3e55c1f8a2b64d9e0c7731aa58e2f6b4d1c09e3a7f25b8c6e41a' as string | null },
	{ id: 'cnt_zv002', code: 'ZV-002', token: '01JD8ZV002CNTR3H8B6YT1MCQE', status: 'active' as 'active' | 'reprinted' | 'revoked', dispatch: 'Packing', batches: ['A-106'], route: 'Pune → Mundra → Antwerp', anchored_at: null as string | null, merkle_root: null as string | null }
];

export const INVENTORY: InventoryItem[] = [
	{ id: 'inv1', code: 'RM-ALI-P1020', name: 'Primary aluminium ingot P1020', category: 'raw_material', qty: 86.4, unit: 't', reorder: 60, unit_value_usd: 2410, tally_id: 'STK-0001', lot: 'HIN-RNK-2611-07' },
	{ id: 'inv2', code: 'RM-SCRAP-PC', name: 'Post-consumer aluminium scrap', category: 'raw_material', qty: 21.7, unit: 't', reorder: 25, unit_value_usd: 1540, tally_id: 'STK-0002', lot: 'SCR-PNQ-2611-03' },
	{ id: 'inv3', code: 'RM-MG-999', name: 'Magnesium 99.9%', category: 'raw_material', qty: 1.9, unit: 't', reorder: 1.2, unit_value_usd: 3900, tally_id: 'STK-0003' },
	{ id: 'inv4', code: 'RM-SI-553', name: 'Silicon metal 553', category: 'raw_material', qty: 0.8, unit: 't', reorder: 1.0, unit_value_usd: 2650, tally_id: 'STK-0004' },
	{ id: 'inv5', code: 'WIP-BIL-6063', name: 'Cast billets 6063 (WIP)', category: 'wip', qty: 14.2, unit: 't', reorder: 0, unit_value_usd: 2550, tally_id: 'STK-0101' },
	{ id: 'inv6', code: 'WIP-SHT-5052', name: 'Rolled sheet coil 5052 (WIP)', category: 'wip', qty: 3.1, unit: 't', reorder: 0, unit_value_usd: 2900, tally_id: 'STK-0102' },
	{ id: 'inv7', code: 'FG-ALB-6063', name: 'Aluminium bars 6063-T6', category: 'finished_good', qty: 18.6, unit: 't', reorder: 8, unit_value_usd: 2860, tally_id: 'STK-0201' },
	{ id: 'inv8', code: 'FG-ALP-6061', name: 'Extruded profiles 6061', category: 'finished_good', qty: 9.8, unit: 't', reorder: 6, unit_value_usd: 3120, tally_id: 'STK-0202' },
	{ id: 'inv9', code: 'FG-ALW-1350', name: 'Aluminium wire 1350', category: 'finished_good', qty: 4.4, unit: 't', reorder: 5, unit_value_usd: 2780, tally_id: 'STK-0203' },
	{ id: 'inv10', code: 'CON-PAL-EU', name: 'Heat-treated EUR pallets (ISPM-15)', category: 'consumable', qty: 140, unit: 'pcs', reorder: 120, unit_value_usd: 14, tally_id: 'STK-0301' },
	{ id: 'inv11', code: 'CON-QR-LBL', name: 'Polyester QR label rolls', category: 'consumable', qty: 6, unit: 'rolls', reorder: 10, unit_value_usd: 38, tally_id: 'STK-0302' }
];

export const MOVEMENTS = [
	{ at: ago(35), kind: 'Issued to line', item: 'RM-ALI-P1020', qty: -4.2, ref: 'A-104', by: 'Vikram Rao' },
	{ at: ago(80), kind: 'Packed into batch', item: 'FG-ALW-1350', qty: -1.9, ref: 'A-106', by: 'Imran Shaikh' },
	{ at: ago(190), kind: 'Goods received', item: 'RM-SCRAP-PC', qty: 6.5, ref: 'GRN-2611-118', by: 'Vikram Rao' },
	{ at: ago(410), kind: 'Issued to line', item: 'RM-MG-999', qty: -0.05, ref: 'A-105', by: 'Sunita Kale' },
	{ at: ago(1300), kind: 'Dispatched', item: 'FG-ALB-6063', qty: -11.5, ref: 'ZV-001', by: 'Imran Shaikh' },
	{ at: ago(2900), kind: 'Goods received', item: 'RM-ALI-P1020', qty: 40, ref: 'GRN-2611-112', by: 'Vikram Rao' }
];

/* ---------- 12 months of Tally vouchers (deterministic) ---------- */
function rng(seed: number) {
	return () => {
		seed = (seed * 1664525 + 1013904223) % 4294967296;
		return seed / 4294967296;
	};
}
export const MONTHS = ['Dec', 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov'];
export const VOUCHERS: Voucher[] = (() => {
	const r = rng(42);
	const out: Voucher[] = [];
	const buyers = ['Rhein Aluminium Import GmbH', 'Nordhavn Metals ApS', 'Lyon Profilés SAS', 'Bharat Fabricators', 'Kolhapur Auto Components', 'Valencia Extrusiones SL'];
	const exportBuyer = (b: string) => !['Bharat Fabricators', 'Kolhapur Auto Components'].includes(b);
	for (let m = 0; m < 12; m++) {
		const year = m === 0 ? 2025 : 2026;
		const month = m === 0 ? 12 : m;
		const growth = 1 + m * 0.035;
		for (let k = 0; k < 9; k++) {
			const b = buyers[Math.floor(r() * buyers.length)];
			const amt = Math.round((14000 + r() * 22000) * growth);
			out.push({ id: `v${m}-${k}`, source: m < 3 ? 'tally_csv' : 'tally_xml', ref: `SAL/${year}${String(month).padStart(2, '0')}/${String(400 + k).padStart(4, '0')}`, type: 'sales', counterparty: b, date: `${year}-${String(month).padStart(2, '0')}-${String(2 + k * 3).padStart(2, '0')}`, amount_usd: amt, cost_usd: Math.round(amt * (0.68 + r() * 0.06)), is_export: exportBuyer(b) });
		}
		for (let k = 0; k < 4; k++) {
			const amt = Math.round((22000 + r() * 18000) * growth);
			out.push({ id: `p${m}-${k}`, source: 'tally_xml', ref: `PUR/${year}${String(month).padStart(2, '0')}/${String(200 + k).padStart(4, '0')}`, type: 'purchase', counterparty: k % 2 ? 'NALCO Angul' : 'Hindalco Renukoot', date: `${year}-${String(month).padStart(2, '0')}-${String(4 + k * 6).padStart(2, '0')}`, amount_usd: amt, cost_usd: amt, is_export: false });
		}
		const util = Math.round((17000 + r() * 3000) * (1 + m * 0.02));
		out.push({ id: `u${m}`, source: 'tally_xml', ref: `UTL/${year}${String(month).padStart(2, '0')}/MSEDCL`, type: 'utility', counterparty: 'MSEDCL', date: `${year}-${String(month).padStart(2, '0')}-05`, amount_usd: util, cost_usd: util, is_export: false });
		const fr = Math.round(5200 + r() * 2400);
		out.push({ id: `f${m}`, source: m > 8 ? 'freightlink' : 'manual', ref: `FRT/${year}${String(month).padStart(2, '0')}/01`, type: 'freight', counterparty: 'Shree Ganesh Roadlines', date: `${year}-${String(month).padStart(2, '0')}-18`, amount_usd: fr, cost_usd: fr, is_export: false });
		const exp = Math.round(9800 + r() * 2200);
		out.push({ id: `e${m}`, source: 'tally_xml', ref: `EXP/${year}${String(month).padStart(2, '0')}/PAY`, type: 'expense', counterparty: 'Payroll & overheads', date: `${year}-${String(month).padStart(2, '0')}-28`, amount_usd: exp, cost_usd: exp, is_export: false });
	}
	return out;
})();

export const ORDERS = {
	sales: [
		{ ref: 'SO-2611-031', customer: 'Rhein Aluminium Import GmbH', product: 'ALB-6063', qty: 22, value: 64800, incoterm: 'FOB JNPT', due: '2026-12-04', status: 'In production', export: true },
		{ ref: 'SO-2611-029', customer: 'Lyon Profilés SAS', product: 'ALP-6061', qty: 14, value: 45200, incoterm: 'CIF Le Havre', due: '2026-12-12', status: 'Scheduled', export: true },
		{ ref: 'SO-2611-027', customer: 'Kolhapur Auto Components', product: 'ALS-5052', qty: 6, value: 20500, incoterm: 'EXW Chakan', due: '2026-11-29', status: 'Ready', export: false },
		{ ref: 'SO-2611-024', customer: 'Nordhavn Metals ApS', product: 'ALW-1350', qty: 9, value: 25400, incoterm: 'FOB Mundra', due: '2026-12-20', status: 'Confirmed', export: true }
	],
	purchases: [
		{ ref: 'PO-2611-118', supplier: 'Hindalco Renukoot', item: 'P1020 ingot', qty: 40, value: 96400, due: '2026-12-01', status: 'In transit' },
		{ ref: 'PO-2611-115', supplier: 'Pune Scrap Aggregators LLP', item: 'PC scrap', qty: 12, value: 18480, due: '2026-11-28', status: 'Confirmed' },
		{ ref: 'PO-2611-109', supplier: 'Ferro Alloys Nashik', item: 'Magnesium 99.9%', qty: 1, value: 3900, due: '2026-12-06', status: 'Awaiting invoice' }
	],
	receivables: [
		{ bucket: '0–30 d', amount: 118400 },
		{ bucket: '31–60 d', amount: 46200 },
		{ bucket: '61–90 d', amount: 12800 },
		{ bucket: '90+ d', amount: 4100 }
	],
	payables: [
		{ bucket: '0–30 d', amount: 102300 },
		{ bucket: '31–60 d', amount: 21900 },
		{ bucket: '61–90 d', amount: 3600 },
		{ bucket: '90+ d', amount: 0 }
	]
};

export const NOTIFICATIONS: Notification[] = [
	{ id: 'n1', category: 'buyer_agency', sender_name: 'Lena Hoffmann', sender_email: 'l.hoffmann@rhein-alu-import.de', org_badge: 'Rhein Aluminium Import · EU Importer', subject: 'RFQ — 40 t 6063-T6 bars, Q1 2027 + CBAM XML', snippet: 'Please quote 40 t of 6063-T6 bars for Q1 delivery to Duisburg. We will need your CBAM communication template and embedded emissions per batch by 15 Dec…', tag: 'CN 76042910', deadline: '2026-12-15T17:00:00+01:00', severity: 'elevated', link: { type: 'rfq', href: '/connect', label: 'Convert to ZENRES RFQ' }, read: false, received_at: ago(18) },
	{ id: 'n2', category: 'buyer_agency', sender_name: 'Arjun Mehta', sender_email: 'arjun@euro-sourcing-agents.com', org_badge: 'Euro Sourcing Agents · Buying house', subject: 'Sample order — extruded profiles for Lyon client', snippet: 'Our French client would like a 500 kg sample of 6061 hollow profiles with container specs and your QR passport link before they place an annual order.', tag: 'CN 76042100', severity: 'normal', link: { type: 'rfq', href: '/connect', label: 'Convert to ZENRES RFQ' }, read: false, received_at: ago(95) },
	{ id: 'n3', category: 'buyer_agency', sender_name: 'Mikkel Sørensen', sender_email: 'ms@nordhavn-metals.dk', org_badge: 'Nordhavn Metals · Distributor', subject: 'Price quote request: EC grade wire 1350', snippet: 'Could you share FOB Mundra pricing for 9 t of 1350 wire and confirm whether emissions data is measured or estimated?', tag: 'CN 76051100', severity: 'normal', read: true, received_at: ago(1440) },
	{ id: 'n4', category: 'deadline', sender_name: 'EU CBAM calendar', sender_email: 'calendar@zenres.example', subject: 'Q4 embedded-emissions data due to Rhein Aluminium Import', snippet: 'Your importer’s declaration window requires supplier emissions data for all Q4 shipments.', deadline: ahead(4), severity: 'elevated', link: { type: 'cbam_report', href: '/carbon/cbam', label: 'Jump to CBAM task' }, read: false, received_at: ago(300) },
	{ id: 'n5', category: 'deadline', sender_name: 'Evidence vault', sender_email: 'vault@zenres.example', subject: 'October electricity bill (MSEDCL) missing', snippet: 'Scope 2 for October is currently estimated. Upload the HT bill to switch to measured.', deadline: ahead(-2), severity: 'critical', link: { type: 'cbam_report', href: '/carbon/inputs', label: 'Upload bill' }, read: false, received_at: ago(2880) },
	{ id: 'n6', category: 'deadline', sender_name: 'Verifier workspace', sender_email: 'verifier@zenres.example', subject: 'CA sign-off pending — Q3 calculation v4', snippet: 'R. Kulkarni & Associates has 3 open comments on precursor allocation.', deadline: ahead(9), severity: 'normal', link: { type: 'audit', href: '/carbon/audit', label: 'Open audit' }, read: true, received_at: ago(4100) },
	{ id: 'n7', category: 'operational_alert', sender_name: 'Warehouse', sender_email: 'system@zenres.example', subject: 'Low stock — Post-consumer scrap below reorder level', snippet: 'RM-SCRAP-PC is at 21.7 t against a 25 t reorder level.', severity: 'elevated', link: { type: 'zone', href: '/dashboard?zone=WH-MAIN', label: 'Inspect zone' }, read: false, received_at: ago(140) },
	{ id: 'n8', category: 'operational_alert', sender_name: 'FreightLink', sender_email: 'freightlink@zenres.example', subject: 'Return load matched — Mumbai → Delhi, 14 t', snippet: 'MH-04-KT-7781 (32 ft MXL) matched with Bhiwandi machine parts. Quote ₹38,500 awaiting your review.', severity: 'normal', link: { type: 'freight', href: '/freight/marketplace', label: 'Review freight quote' }, read: false, received_at: ago(52) },
	{ id: 'n9', category: 'operational_alert', sender_name: 'Tally connector', sender_email: 'system@zenres.example', subject: 'Sync mismatch — 2 vouchers need review', snippet: 'SAL/202611/0407 amount differs between Tally (₹21,48,000) and the last import.', severity: 'normal', link: { type: 'tally', href: '/integrations/tally', label: 'View sync log' }, read: true, received_at: ago(720) }
];

export const FREIGHT = {
	trucks: [
		{ id: 'tr1', reg: 'MH-04-KT-7781', carrier: 'Shree Ganesh Roadlines', vehicle: '32 ft MXL container', payload: 21, origin: 'Mumbai', dest: 'Delhi', corridor: 'BOM-DEL', window: ['2026-11-27T06:00', '2026-11-28T20:00'], detour_km: 18, quote_inr: 38500, reliability: 4.8, verified: true, docs_valid: true },
		{ id: 'tr2', reg: 'HR-55-AL-2290', carrier: 'Haryana Freight Movers', vehicle: '32 ft SXL container', payload: 16, origin: 'Mumbai', dest: 'Delhi', corridor: 'BOM-DEL', window: ['2026-11-27T12:00', '2026-11-29T12:00'], detour_km: 42, quote_inr: 35200, reliability: 4.5, verified: true, docs_valid: true },
		{ id: 'tr3', reg: 'PB-10-GH-4412', carrier: 'Punjab Express Logistics', vehicle: '19 ft closed body', payload: 10, origin: 'Delhi', dest: 'Ludhiana', corridor: 'DEL-LDH', window: ['2026-11-26T18:00', '2026-11-27T10:00'], detour_km: 6, quote_inr: 14800, reliability: 4.9, verified: true, docs_valid: true },
		{ id: 'tr4', reg: 'MH-12-RT-0915', carrier: 'Deccan Carriers', vehicle: '20 ft open body', payload: 9, origin: 'Mumbai', dest: 'Delhi', corridor: 'BOM-DEL', window: ['2026-11-30T06:00', '2026-12-01T06:00'], detour_km: 64, quote_inr: 29900, reliability: 4.1, verified: false, docs_valid: false }
	],
	loads: [
		{ id: 'ld1', shipper: 'Bhiwandi Precision Parts', origin: 'Mumbai', dest: 'Delhi', corridor: 'BOM-DEL', weight: 14, cargo: 'Packaged machine parts', packaging: 'Crated', window: ['2026-11-27T08:00', '2026-11-28T18:00'], budget_inr: 40000 },
		{ id: 'ld2', shipper: 'Okhla Garment Exports', origin: 'Delhi', dest: 'Ludhiana', corridor: 'DEL-LDH', weight: 7, cargo: 'Textile cartons', packaging: 'Cartons', window: ['2026-11-26T20:00', '2026-11-27T12:00'], budget_inr: 16000 }
	],
	shipments: [
		{ id: 'sh1', ref: 'FL-2611-014', route: 'Chakan → JNPT Nhava Sheva', corridor: 'PNQ-JNPT', carrier: 'Shree Ganesh Roadlines', vehicle: 'MH-14-HG-3320', status: 'in_transit', gps_consent: true, freight_inr: 21400, fee_pct: 2, baseline_l: 180, matched_l: 140, pickup_scanned: true, delivery_scanned: false, progress: 0.62, eta: ahead(0.2), container: 'ZV-001' },
		{ id: 'sh2', ref: 'FL-2611-011', route: 'Delhi → Ludhiana', corridor: 'DEL-LDH', carrier: 'Punjab Express Logistics', vehicle: 'PB-10-GH-4412', status: 'delivered', gps_consent: false, freight_inr: 14800, fee_pct: 2, baseline_l: 64, matched_l: 51, pickup_scanned: true, delivery_scanned: true, progress: 1, eta: ago(1500), container: null },
		{ id: 'sh3', ref: 'FL-2611-016', route: 'Mumbai → Delhi', corridor: 'BOM-DEL', carrier: 'Haryana Freight Movers', vehicle: 'HR-55-AL-2290', status: 'confirmed', gps_consent: false, freight_inr: 35200, fee_pct: 2, baseline_l: 410, matched_l: 322, pickup_scanned: false, delivery_scanned: false, progress: 0, eta: ahead(2), container: null }
	],
	transporters: [
		{ name: 'Shree Ganesh Roadlines', city: 'Navi Mumbai', vehicles: 14, verified: true, reliability: 4.8, on_time: 96, docs: { rc: true, insurance: true, permit: true, kyc: true } },
		{ name: 'Punjab Express Logistics', city: 'Ludhiana', vehicles: 9, verified: true, reliability: 4.9, on_time: 98, docs: { rc: true, insurance: true, permit: true, kyc: true } },
		{ name: 'Haryana Freight Movers', city: 'Gurugram', vehicles: 22, verified: true, reliability: 4.5, on_time: 91, docs: { rc: true, insurance: true, permit: true, kyc: true } },
		{ name: 'Deccan Carriers', city: 'Pune', vehicles: 5, verified: false, reliability: 4.1, on_time: 84, docs: { rc: true, insurance: false, permit: true, kyc: false } }
	]
};

export const CONNECT = {
	suppliers: [
		{ name: 'Deccan Alloy Works', city: 'Pune', country: 'IN', sector: 'Aluminium', cn: ['76042910', '76042100', '76061220'], capacity: '335 t/mo', tier: 'Document-Reviewed', intensity: 5.5, certs: ['ISO 9001', 'ISO 14001'] },
		{ name: 'Gujarat Steel Tubes', city: 'Ahmedabad', country: 'IN', sector: 'Iron & Steel', cn: ['73063019', '73066190'], capacity: '1,200 t/mo', tier: 'Independently Verified', intensity: 2.1, certs: ['ISO 9001', 'EN 10219'] },
		{ name: 'Coimbatore Castings', city: 'Coimbatore', country: 'IN', sector: 'Iron & Steel', cn: ['73251000'], capacity: '420 t/mo', tier: 'Self-Declared', intensity: 1.9, certs: ['IATF 16949'] },
		{ name: 'Vizag Aluminium Foils', city: 'Visakhapatnam', country: 'IN', sector: 'Aluminium', cn: ['76071190'], capacity: '260 t/mo', tier: 'Document-Reviewed', intensity: 6.8, certs: ['ISO 9001'] },
		{ name: 'Rajasthan Cement Works', city: 'Chittorgarh', country: 'IN', sector: 'Cement', cn: ['25232900'], capacity: '30,000 t/mo', tier: 'Independently Verified', intensity: 0.71, certs: ['ISO 14001', 'BIS'] }
	],
	rfqs: [
		{ id: 'RFQ-2611-008', buyer: 'Rhein Aluminium Import GmbH', product: 'ALB-6063', qty: 40, incoterm: 'FOB JNPT', status: 'open', quotes: [{ supplier: 'Deccan Alloy Works', price: 2985, lead: 21, intensity: 5.5, tier: 'Document-Reviewed' }] },
		{ id: 'RFQ-2611-006', buyer: 'Lyon Profilés SAS', product: 'ALP-6061', qty: 14, incoterm: 'CIF Le Havre', status: 'negotiating', quotes: [{ supplier: 'Deccan Alloy Works', price: 3290, lead: 25, intensity: 5.89, tier: 'Document-Reviewed' }, { supplier: 'Vizag Aluminium Foils', price: 3180, lead: 32, intensity: 6.8, tier: 'Document-Reviewed' }] },
		{ id: 'RFQ-2610-031', buyer: 'Nordhavn Metals ApS', product: 'ALW-1350', qty: 9, incoterm: 'FOB Mundra', status: 'awarded', quotes: [{ supplier: 'Deccan Alloy Works', price: 2860, lead: 18, intensity: 6.24, tier: 'Self-Declared' }] }
	]
};

export const IMPORTER_SUPPLIERS = [
	['Deccan Alloy Works', 'Pune, IN', '76042910', 'verified', 5.5, 18.0],
	['Vizag Aluminium Foils', 'Visakhapatnam, IN', '76071190', 'submitted', 6.8, 9.2],
	['Gujarat Steel Tubes', 'Ahmedabad, IN', '73063019', 'verified', 2.1, 64.0],
	['Coimbatore Castings', 'Coimbatore, IN', '73251000', 'pending_evidence', 1.9, 22.5],
	['Jindal Wire Products', 'Hisar, IN', '72172000', 'submitted', 2.4, 31.0],
	['Kutch Fertilisers', 'Gandhidham, IN', '31021010', 'onboarding', 3.1, 40.0],
	['Bhilai Structurals', 'Bhilai, IN', '72163100', 'verified', 2.6, 55.0],
	['Hosur Precision Forge', 'Hosur, IN', '73269099', 'pending_evidence', 2.2, 12.4],
	['Nagpur Alloy Castings', 'Nagpur, IN', '76169990', 'submitted', 7.4, 6.8],
	['Raipur Sponge Iron', 'Raipur, IN', '72031000', 'pending_evidence', 3.6, 48.0],
	['Ludhiana Fasteners', 'Ludhiana, IN', '73181500', 'verified', 2.0, 15.5],
	['Chennai Cast Aluminium', 'Chennai, IN', '76169910', 'onboarding', 6.1, 7.2],
	['Rajkot Brass & Alu', 'Rajkot, IN', '76090000', 'invited', null, 4.0],
	['Odisha Ferro Chrome', 'Jajpur, IN', '72024100', 'invited', null, 11.0],
	['Kolkata Wire Ropes', 'Howrah, IN', '73121090', 'submitted', 2.3, 9.8]
].map(([name, city, cn, status, intensity, tonnes], i) => ({
	id: `sup${i + 1}`,
	name: name as string,
	city: city as string,
	cn: cn as string,
	status: status as 'invited' | 'onboarding' | 'pending_evidence' | 'submitted' | 'verified',
	intensity: intensity as number | null,
	tonnes: tonnes as number,
	last_reminder: i % 3 === 0 ? ago(1440 * (i + 1)) : null
}));

export const TEAM = [
	{ name: 'Ananya Deshpande', email: 'ananya@deccanalloy.in', role: 'owner', last: ago(2) },
	{ name: 'Prakash Iyer', email: 'accounts@deccanalloy.in', role: 'finance', last: ago(64) },
	{ name: 'Rakesh Patil', email: 'rakesh.p@deccanalloy.in', role: 'production', last: ago(12) },
	{ name: 'Meera Joshi', email: 'compliance@deccanalloy.in', role: 'compliance', last: ago(230) },
	{ name: 'Shree Ganesh Roadlines', email: 'ops@shreeganeshroadlines.in', role: 'logistics_partner', last: ago(1600) },
	{ name: 'Lena Hoffmann', email: 'l.hoffmann@rhein-alu-import.de', role: 'buyer', last: ago(3000) },
	{ name: 'R. Kulkarni & Associates', email: 'audit@rkulkarni-ca.in', role: 'verifier', last: ago(4300) }
];

export const ROLE_LABEL: Record<string, string> = {
	owner: 'MSME Owner / Admin',
	finance: 'Finance / Accounts',
	production: 'Production Manager',
	compliance: 'Sustainability / Compliance',
	logistics_partner: 'Logistics Partner',
	buyer: 'Buyer',
	verifier: 'External Verifier / CA'
};

export const ACTIVITY = [
	{ id: 'act1', period: '2026-10', type: 'electricity', mode: 'tally_csv', qty: 412800, unit: 'kWh', ef: 0.000727, rule: 'CEA-v20 / IR 2025/2621', file: 'MSEDCL_HT_Oct26.pdf', quality: 'estimated' as DataQuality },
	{ id: 'act2', period: '2026-10', type: 'fuel', mode: 'manual', qty: 28400, unit: 'L diesel', ef: 0.00268, rule: 'IPCC 2006 / IR 2025/2621', file: 'DG_log_Oct26.xlsx', quality: 'measured' as DataQuality },
	{ id: 'act3', period: '2026-10', type: 'fuel', mode: 'manual', qty: 61200, unit: 'Sm³ natural gas', ef: 0.00202, rule: 'IPCC 2006 / IR 2025/2621', file: 'MNGL_Oct26.pdf', quality: 'measured' as DataQuality },
	{ id: 'act4', period: '2026-10', type: 'precursor', mode: 'manual', qty: 118, unit: 't P1020 ingot', ef: 3.9, rule: 'Supplier EPD / IR 2025/2621', file: 'Hindalco_EPD_2026.pdf', quality: 'supplier_reported' as DataQuality },
	{ id: 'act5', period: '2026-09', type: 'electricity', mode: 'tally_csv', qty: 398100, unit: 'kWh', ef: 0.000727, rule: 'CEA-v20 / IR 2025/2621', file: 'MSEDCL_HT_Sep26.pdf', quality: 'measured' as DataQuality },
	{ id: 'act6', period: '2026-09', type: 'process', mode: 'manual', qty: 3.2, unit: 't anode consumption', ef: 1.62, rule: 'IR 2025/2621 Annex III', file: 'Process_log_Sep26.xlsx', quality: 'measured' as DataQuality }
];

export const VERIFIERS = [
	{ name: 'R. Kulkarni & Associates', city: 'Pune', type: 'Chartered Accountants', accredited: 'CBAM-ready verifier network', rating: 4.9, referral: true },
	{ name: 'TÜV Rheinland India', city: 'Bengaluru', type: 'Accredited verification body', accredited: 'EN ISO 14065', rating: 4.8, referral: true },
	{ name: 'GreenLedger Assurance LLP', city: 'Mumbai', type: 'Sustainability assurance', accredited: 'ISO 14064-3', rating: 4.6, referral: false }
];

export const OPS = {
	approvals: [
		{ id: 'apr1', kind: 'Transporter KYC', name: 'Deccan Carriers', detail: 'RC ✓ · Insurance ✗ (expired 2026-10-31) · Permit ✓', submitted: ago(300) },
		{ id: 'apr2', kind: 'Vehicle permit', name: 'MH-12-RT-0915', detail: 'National permit renewal uploaded', submitted: ago(820) },
		{ id: 'apr3', kind: 'MSME verification', name: 'Nagpur Alloy Castings', detail: 'GSTIN 27AAFCN9921M1ZQ · Udyam certificate', submitted: ago(1600) }
	],
	disputes: [
		{ id: 'dsp1', ref: 'FL-2610-088', parties: 'Okhla Garment Exports ↔ Punjab Express', issue: 'Short delivery — 2 cartons', amount: '₹6,400', status: 'Evidence requested' },
		{ id: 'dsp2', ref: 'FL-2610-071', parties: 'Bhiwandi Precision ↔ Deccan Carriers', issue: 'Detention charges disputed', amount: '₹3,000', status: 'Under review' }
	],
	margins: [
		{ tier: 'MSME Starter', price: 99, annual_eq: 82.5, ceiling: 19.8, ceiling_pct: 80, annual_ceiling: 16.5, actual: 17.2, breakdown: { cloud_ocr: 5.1, support: 6.4, compliance_ops: 4.2, freight_maps: 1.5 } },
		{ tier: 'MSME Growth', price: 249, annual_eq: 207.5, ceiling: 49.8, ceiling_pct: 80, annual_ceiling: 41.5, actual: 44.6, breakdown: { cloud_ocr: 14.9, support: 12.2, compliance_ops: 9.8, freight_maps: 7.7 } },
		{ tier: 'EU Importer Hub', price: 499, annual_eq: 415.83, ceiling: 74.85, ceiling_pct: 85, annual_ceiling: 62.38, actual: 66.1, breakdown: { cloud_ocr: 19.4, support: 21.8, compliance_ops: 15.2, freight_maps: 9.7 } }
	],
	kpis: [
		['Match rate', '68%', '+6 pts'],
		['Booking conversion', '41%', '+3 pts'],
		['Empty km %', '22%', '−9 pts'],
		['Payload utilisation', '81%', '+5 pts'],
		['On-time delivery', '94%', '+1 pt'],
		['Freight savings', '₹4.2 L', 'pilot to date'],
		['Pilot time-to-data-pack', '11 days', '−4 days']
	]
};
