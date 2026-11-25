import { describe, it, expect } from 'vitest';
import { specificEmbedded, compareDefaultVsActual, article9Deduction, avoidedFuel, buildChain, verifyChain, canonical, capexScenario, activityEmissions } from './carbon';

const bar = { direct_tco2e_t: 0.62, indirect_tco2e_t: 1.71, precursor_tco2e_t: 3.17, default_tco2e_t: 7.9 };

describe('CBAM math', () => {
	it('sums direct, indirect and precursor emissions deterministically', () => {
		expect(specificEmbedded(bar)).toBe(5.5);
		expect(specificEmbedded(bar, { includePrecursors: false })).toBe(2.33);
	});
	it('compares default values against measured actuals', () => {
		const r = compareDefaultVsActual(100, bar, 76);
		expect(r.defaultEur).toBe(60040);
		expect(r.actualEur).toBe(41800);
		expect(r.savingsEur).toBe(18240);
	});
	it('caps Article 9 deductions at the certificate price', () => {
		expect(article9Deduction(10, 12, 76)).toBe(120);
		expect(article9Deduction(10, 500, 76)).toBe(760);
		expect(article9Deduction(10, -3, 76)).toBe(0);
	});
	it('converts activity data with tCO2e-per-unit factors', () => {
		expect(activityEmissions(412800, 0.000727)).toBe(300.1056);
	});
});

describe('FreightLink avoided fuel', () => {
	it('matches the blueprint 180 L vs 140 L example and stays a separate logistics estimate', () => {
		const r = avoidedFuel(180, 140);
		expect(r.litres).toBe(40);
		expect(r.kgCO2).toBe(107.2);
		expect(r.scope).toBe('logistics_estimate');
	});
	it('never reports negative savings', () => {
		expect(avoidedFuel(100, 130).litres).toBe(0);
	});
});

describe('tamper-evident hash chain', () => {
	it('produces canonical JSON with sorted keys', () => {
		expect(canonical({ b: 1, a: [2, { d: 3, c: 4 }] })).toBe('{"a":[2,{"c":4,"d":3}],"b":1}');
	});
	it('detects edits anywhere in the chain', async () => {
		const chain = await buildChain([{ id: 1, kwh: 100 }, { id: 2, kwh: 200 }, { id: 3, kwh: 300 }]);
		expect(await verifyChain(chain)).toBe(-1);
		const tampered = chain.map((c) => ({ ...c }));
		tampered[1].kwh = 199;
		expect(await verifyChain(tampered)).toBe(1);
	});
});

describe('capex ROI simulator', () => {
	it('computes payback and IRR', () => {
		const r = capexScenario({ capexUsd: 300000, annualSavingsUsd: 60000, annualAvoidedT: 400, years: 10, depreciationPct: 40, taxRatePct: 25, certPriceEur: 76, exportShare: 0.7 });
		expect(r.payback).not.toBeNull();
		expect(r.payback!).toBeLessThan(5);
		expect(r.irr!).toBeGreaterThan(10);
	});
});
