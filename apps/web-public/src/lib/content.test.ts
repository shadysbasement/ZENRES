import { describe, it, expect } from 'vitest';
import { TIERS, COMPARISON, cbamExposure, formatPrice, CORRIDORS, FREIGHT_STAGES, JOURNEY, PILLARS, DISCLAIMERS } from './content';

describe('commercial terms match the blueprint', () => {
	it('prices and annual 2-months-free', () => {
		expect(TIERS.map((t) => [t.monthly, t.annual])).toEqual([
			[99, 990],
			[249, 2490],
			[499, 4990]
		]);
		for (const t of TIERS) expect(t.annual).toBe(t.monthly * 10);
	});
	it('has a 14-row comparison matrix', () => {
		expect(COMPARISON).toHaveLength(14);
	});
	it('formats indicative currencies', () => {
		expect(formatPrice(249, 'USD')).toBe('$249');
		expect(formatPrice(249, 'INR')).toBe('₹21,912');
	});
});

describe('landing content', () => {
	it('has 4 pillars, 8 journey steps and 8 FreightLink stages', () => {
		expect(PILLARS).toHaveLength(4);
		expect(JOURNEY).toHaveLength(8);
		expect(FREIGHT_STAGES).toHaveLength(8);
	});
	it('uses the blueprint 180 L vs 140 L backhaul example on Delhi–Mumbai', () => {
		const c = CORRIDORS.find((x) => x.id === 'DEL-BOM')!;
		expect(c.baselineFuelL - c.matchedFuelL).toBe(40);
	});
	it('illustrative CBAM exposure scales linearly', () => {
		const a = cbamExposure(100);
		const b = cbamExposure(200);
		expect(b.deltaEur).toBeCloseTo(a.deltaEur * 2);
		expect(a.defaultEur).toBeGreaterThan(a.actualEur);
	});
	it('carries the mandatory disclaimers', () => {
		expect(DISCLAIMERS.reporting).toMatch(/not legal, tax or accounting advice/);
		expect(DISCLAIMERS.anchoring).toMatch(/not source-input truthfulness/);
		expect(DISCLAIMERS.buyers).toMatch(/do not imply confirmed partnerships/);
	});
});
