import { describe, it, expect } from 'vitest';
import { rankMatches, rejected, createBooking, type Truck, type Load } from './freight';
import { FREIGHT } from './demo/seed';

const trucks = FREIGHT.trucks.map((t) => ({ ...t, window: t.window as [string, string] })) as Truck[];
const load: Load = { ...FREIGHT.loads[0], window: FREIGHT.loads[0].window as [string, string] };

describe('FreightLink matcher', () => {
	it('ranks the Mumbai→Delhi return load with MH-04-KT-7781 first', () => {
		const m = rankMatches(trucks, load);
		expect(m[0].truck.reg).toBe('MH-04-KT-7781');
		expect(m.map((x) => x.truck.corridor).every((c) => c === 'BOM-DEL')).toBe(true);
	});
	it('applies hard filters before ranking', () => {
		const r = rejected(trucks, load);
		const reasons = Object.fromEntries(r.map((x) => [x.truck.reg, x.reason]));
		expect(reasons['PB-10-GH-4412']).toBe('corridor');
		expect(reasons['MH-12-RT-0915']).toBe('window');
	});
	it('ignores paid promotion when ranking', () => {
		const promoted = trucks.map((t) => (t.reg === 'HR-55-AL-2290' ? { ...t, promoted: true } : t));
		expect(rankMatches(promoted, load)[0].truck.reg).toBe(rankMatches(trucks, load)[0].truck.reg);
	});
	it('never auto-books: explicit authorised acceptance is required', () => {
		const m = rankMatches(trucks, load)[0];
		expect(() => createBooking(m, { accepted: false, feePct: 2 })).toThrow('explicit_acceptance_required');
		const b = createBooking(m, { accepted: true, authorisedBy: 'Ananya Deshpande', feePct: 2 });
		expect(b.disclosed_fee_inr).toBe(770);
		expect(b.total_inr).toBe(39270);
	});
});
