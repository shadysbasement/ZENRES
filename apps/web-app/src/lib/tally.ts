import type { Voucher } from './demo/seed';

export const TALLY_FIELDS = [
	{ key: 'date', label: 'Voucher date', match: /^(date|voucher ?date|vch ?date)$/i },
	{ key: 'ref', label: 'Voucher no. (stable ref)', match: /(voucher ?(no|number)|vch ?no|ref|guid)/i },
	{ key: 'type', label: 'Voucher type', match: /(voucher ?type|vch ?type|type)/i },
	{ key: 'counterparty', label: 'Party ledger', match: /(party|ledger|particulars|customer|supplier)/i },
	{ key: 'amount', label: 'Amount (₹)', match: /(amount|debit|credit|value)/i },
	{ key: 'export', label: 'Export flag', match: /(export|foreign|lut)/i }
] as const;

export type ColumnMap = Record<string, string>;

/** Minimal RFC-4180 CSV parser (quoted fields, escaped quotes, CRLF). */
export function parseTallyCsv(text: string): { headers: string[]; rows: string[][] } {
	const out: string[][] = [];
	let row: string[] = [];
	let field = '';
	let q = false;
	for (let i = 0; i < text.length; i++) {
		const c = text[i];
		if (q) {
			if (c === '"' && text[i + 1] === '"') {
				field += '"';
				i++;
			} else if (c === '"') q = false;
			else field += c;
		} else if (c === '"') q = true;
		else if (c === ',') {
			row.push(field.trim());
			field = '';
		} else if (c === '\n' || c === '\r') {
			if (c === '\r' && text[i + 1] === '\n') i++;
			row.push(field.trim());
			if (row.some((x) => x !== '')) out.push(row);
			row = [];
			field = '';
		} else field += c;
	}
	if (field || row.length) {
		row.push(field.trim());
		if (row.some((x) => x !== '')) out.push(row);
	}
	return { headers: out[0] ?? [], rows: out.slice(1) };
}

export function autoMap(headers: string[]): ColumnMap {
	const used = new Set<string>();
	const map: ColumnMap = {};
	for (const h of headers) {
		const f = TALLY_FIELDS.find((f) => !used.has(f.key) && f.match.test(h.trim()));
		map[h] = f?.key ?? '';
		if (f) used.add(f.key);
	}
	return map;
}

const TYPE_MAP: Record<string, Voucher['type']> = { sales: 'sales', sale: 'sales', purchase: 'purchase', payment: 'expense', journal: 'expense', receipt: 'sales', utility: 'utility', freight: 'freight' };

/**
 * Idempotent upsert keyed on (source, external_ref) — mirrors the Postgres
 * UNIQUE (org_id, source, external_ref) constraint used by the core API.
 */
export function upsertVouchers(existing: Voucher[], headers: string[], rows: string[][], map: ColumnMap, source: Voucher['source'], inrPerUsd = 88) {
	const col = (k: string) => headers.findIndex((h) => map[h] === k);
	const ci = { date: col('date'), ref: col('ref'), type: col('type'), party: col('counterparty'), amount: col('amount'), exp: col('export') };
	const seen = new Set(existing.map((v) => `${v.source}|${v.ref}`));
	const vouchers = [...existing];
	let inserted = 0,
		duplicates = 0,
		errors = 0;
	for (const r of rows) {
		const ref = ci.ref >= 0 ? r[ci.ref] : '';
		const amountInr = Number((ci.amount >= 0 ? r[ci.amount] : '').replace(/[₹,\s]/g, ''));
		if (!ref || !Number.isFinite(amountInr)) {
			errors++;
			continue;
		}
		const key = `${source}|${ref}`;
		if (seen.has(key)) {
			duplicates++;
			continue;
		}
		seen.add(key);
		const t = (ci.type >= 0 ? r[ci.type] : 'sales').toLowerCase();
		const amount = Math.round(Math.abs(amountInr) / inrPerUsd);
		vouchers.push({
			id: `imp-${ref}`,
			source,
			ref,
			type: TYPE_MAP[t] ?? 'expense',
			counterparty: ci.party >= 0 ? r[ci.party] : 'Unknown',
			date: normaliseDate(ci.date >= 0 ? r[ci.date] : ''),
			amount_usd: amount,
			cost_usd: Math.round(amount * 0.71),
			is_export: ci.exp >= 0 ? /^(y|yes|true|1)$/i.test(r[ci.exp]) : false
		});
		inserted++;
	}
	return { vouchers, inserted, duplicates, errors };
}

export function normaliseDate(s: string): string {
	const m = s.match(/^(\d{1,2})[-/.](\d{1,2}|[A-Za-z]{3})[-/.](\d{2,4})$/);
	if (!m) return /^\d{4}-\d{2}-\d{2}$/.test(s) ? s : '2026-11-26';
	const months = ['jan', 'feb', 'mar', 'apr', 'may', 'jun', 'jul', 'aug', 'sep', 'oct', 'nov', 'dec'];
	const mo = isNaN(Number(m[2])) ? months.indexOf(m[2].toLowerCase()) + 1 : Number(m[2]);
	const y = m[3].length === 2 ? `20${m[3]}` : m[3];
	return `${y}-${String(mo).padStart(2, '0')}-${m[1].padStart(2, '0')}`;
}

export const SAMPLE_TALLY_CSV = `Date,Voucher No,Voucher Type,Particulars,Amount,Export
21-Nov-2026,SAL/202611/0451,Sales,Rhein Aluminium Import GmbH,"21,84,000",Yes
21-Nov-2026,SAL/202611/0452,Sales,Kolhapur Auto Components,"6,12,500",No
22-Nov-2026,PUR/202611/0231,Purchase,Hindalco Renukoot,"18,40,000",No
23-Nov-2026,SAL/202611/0453,Sales,Lyon Profilés SAS,"14,96,000",Yes
24-Nov-2026,PAY/202611/0102,Payment,MSEDCL,"15,20,400",No
25-Nov-2026,SAL/202611/0454,Sales,Nordhavn Metals ApS,"9,02,000",Yes
`;
