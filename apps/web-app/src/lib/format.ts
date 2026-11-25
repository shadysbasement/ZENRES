import { USD_INR, USD_EUR } from './carbon';
export type Ccy = 'USD' | 'INR' | 'EUR';
export function money(usd: number, ccy: Ccy = 'USD', opts: { compact?: boolean; decimals?: number } = {}): string {
	const v = ccy === 'INR' ? usd * USD_INR : ccy === 'EUR' ? usd * USD_EUR : usd;
	const sym = ccy === 'INR' ? '₹' : ccy === 'EUR' ? '€' : '$';
	if (opts.compact) {
		if (ccy === 'INR') {
			if (Math.abs(v) >= 1e7) return `${sym}${(v / 1e7).toFixed(2)} Cr`;
			if (Math.abs(v) >= 1e5) return `${sym}${(v / 1e5).toFixed(1)} L`;
		}
		if (Math.abs(v) >= 1e6) return `${sym}${(v / 1e6).toFixed(2)}M`;
		if (Math.abs(v) >= 1e3) return `${sym}${(v / 1e3).toFixed(1)}k`;
	}
	return `${sym}${v.toLocaleString(ccy === 'INR' ? 'en-IN' : 'en-US', { maximumFractionDigits: opts.decimals ?? 0, minimumFractionDigits: opts.decimals ?? 0 })}`;
}
export const num = (v: number, d = 0) => v.toLocaleString('en-US', { maximumFractionDigits: d, minimumFractionDigits: d });
export const pct = (v: number, d = 1) => `${v >= 0 ? '+' : ''}${v.toFixed(d)}%`;
export function timeAgo(isoStr: string, now = Date.now()): string {
	const s = Math.round((now - new Date(isoStr).getTime()) / 1000);
	if (s < 60) return 'just now';
	if (s < 3600) return `${Math.floor(s / 60)}m ago`;
	if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
	return `${Math.floor(s / 86400)}d ago`;
}
export function dueIn(isoStr: string, now = Date.now()): { label: string; tone: 'crimson' | 'terracotta' | 'muted' } {
	const d = Math.ceil((new Date(isoStr).getTime() - now) / 86400000);
	if (d < 0) return { label: 'Overdue', tone: 'crimson' };
	if (d === 0) return { label: 'Due today', tone: 'crimson' };
	if (d <= 7) return { label: `Due in ${d}d`, tone: 'terracotta' };
	return { label: `Due in ${d}d`, tone: 'muted' };
}
export const fmtDate = (d: string) => new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
export const QUALITY_LABEL: Record<string, string> = { measured: 'Measured', supplier_reported: 'Supplier-Reported', estimated: 'Estimated', verified: 'Independently Verified' };
export const QUALITY_PILL: Record<string, string> = { measured: 'pill-emerald', supplier_reported: 'pill-cobalt', estimated: 'pill-terracotta', verified: 'pill-dark' };
