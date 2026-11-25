import { describe, it, expect } from 'vitest';
import { buildManifest } from './manifest';
import { evaluateSubtleHint, HINT_RULES, HINT_COOLDOWN_MS } from './hintEngine';

const starter = buildManifest('starter');
const growth = buildManifest('growth');
const hub = buildManifest('importer_hub');

describe('entitlement manifest', () => {
	it('never exposes higher-tier modules to Starter', () => {
		expect(starter.visible_modules).not.toContain('freight/marketplace');
		expect(starter.visible_modules).not.toContain('carbon/simulator');
		expect(starter.visible_modules).not.toContain('importer-hub');
		expect(starter.visible_modules).toContain('carbon/cbam');
	});
	it('reveals FreightLink on Growth and the supplier portal on Importer Hub', () => {
		expect(growth.visible_modules).toContain('freight/marketplace');
		expect(growth.visible_modules).not.toContain('freight/enterprise');
		expect(hub.visible_modules).toContain('importer-hub');
		expect(hub.visible_modules).toContain('freight/enterprise');
	});
	it('encodes blueprint limits', () => {
		expect(starter.limits).toMatchObject({ max_plants: 1, max_cn_skus: 5 });
		expect(growth.limits).toMatchObject({ max_plants: 3, max_cn_skus: null });
		expect(hub.limits.included_suppliers).toBe(15);
	});
});

describe('subtle hint engine', () => {
	it('defines exactly 15 trigger scenarios', () => {
		expect(HINT_RULES).toHaveLength(15);
	});
	it('whispers FreightLink when a Starter user searches for freight', () => {
		const h = evaluateSubtleHint(starter, { kind: 'search_query', payload: 'find a return load' });
		expect(h?.code).toBe('starter.freight_search');
		expect(h?.suggested_tier).toBe('growth');
		expect(h?.learn_more).toBe('/settings/billing?focus=freightlink_core');
	});
	it('stays silent for Growth users searching freight (already entitled)', () => {
		expect(evaluateSubtleHint(growth, { kind: 'search_query', payload: 'truck' })).toBeNull();
	});
	it('fires on the 6th SKU only', () => {
		expect(evaluateSubtleHint(starter, { kind: 'limit_hit', payload: 'add_sku', count: 4 })).toBeNull();
		expect(evaluateSubtleHint(starter, { kind: 'limit_hit', payload: 'add_sku', count: 5 })?.code).toBe('starter.sixth_sku');
	});
	it('respects the 14-day cooldown', () => {
		const now = Date.now();
		const t = { kind: 'search_query' as const, payload: 'solar capex' };
		expect(evaluateSubtleHint(starter, t, { 'starter.capex': now - 1000 }, now)).toBeNull();
		expect(evaluateSubtleHint(starter, t, { 'starter.capex': now - HINT_COOLDOWN_MS - 1 }, now)?.code).toBe('starter.capex');
	});
	it('discloses the $25 extra supplier seat on Importer Hub', () => {
		const h = evaluateSubtleHint(hub, { kind: 'limit_hit', payload: 'invite_supplier', count: 15 });
		expect(h?.suggested_tier).toBe('expansion_addon');
		expect(h?.message).toContain('$25/month');
	});
	it('suggests Importer Hub for consolidated XML on Growth', () => {
		expect(evaluateSubtleHint(growth, { kind: 'entity_pattern', payload: 'consolidated_xml' })?.suggested_tier).toBe('importer_hub');
	});
});
