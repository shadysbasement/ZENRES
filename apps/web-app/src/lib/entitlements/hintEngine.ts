import type { EntitlementManifest, SubscriptionTier, Capability } from './manifest';

export type TriggerKind = 'limit_hit' | 'search_query' | 'manual_effort' | 'entity_pattern';
export interface HintTrigger {
	kind: TriggerKind;
	payload: string;
	/** numeric context, e.g. current plant count */
	count?: number;
}
export interface SubtleHint {
	code: string;
	message: string;
	suggested_tier: SubscriptionTier | 'expansion_addon';
	capability: Capability | 'extra_supplier_seat';
	learn_more: string;
}

interface HintRule {
	code: string;
	tier: SubscriptionTier;
	target: SubtleHint['suggested_tier'];
	capability: SubtleHint['capability'];
	message: string;
	limit?: boolean;
	match: (t: HintTrigger) => boolean;
}

const q = (re: RegExp) => (t: HintTrigger) => t.kind === 'search_query' && re.test(t.payload.toLowerCase());
const is = (kind: TriggerKind, payload: string, pred?: (t: HintTrigger) => boolean) => (t: HintTrigger) =>
	t.kind === kind && t.payload === payload && (!pred || pred(t));

/** Exhaustive trigger-to-whisper matrix (15 scenarios). Microcopy is intentionally calm. */
export const HINT_RULES: HintRule[] = [
	{
		code: 'starter.second_plant',
		limit: true,
		tier: 'starter',
		target: 'growth',
		capability: 'core_erp_tally_csv',
		message: 'Your workspace covers 1 manufacturing plant. Adding a second facility is supported seamlessly on MSME Growth.',
		match: is('limit_hit', 'add_plant', (t) => (t.count ?? 0) >= 1)
	},
	{
		code: 'starter.sixth_sku',
		limit: true,
		tier: 'starter',
		target: 'growth',
		capability: 'precursor_tracking',
		message: "You've mapped all 5 CN-code SKUs on Starter. Unlimited SKUs and precursor tracking open automatically on Growth.",
		match: is('limit_hit', 'add_sku', (t) => (t.count ?? 0) >= 5)
	},
	{
		code: 'starter.ocr',
		tier: 'starter',
		target: 'growth',
		capability: 'ocr_ingestion',
		message: 'Prefer automatic line-item extraction? OCR bill capture and WhatsApp floor-bot ingestion are built into Growth.',
		match: is('manual_effort', 'bill_manual_fields', (t) => (t.count ?? 0) >= 3)
	},
	{
		code: 'starter.freight_search',
		tier: 'starter',
		target: 'growth',
		capability: 'freightlink_core',
		message: 'Looking to match return truck capacity or reduce empty km? FreightLink backhaul matching is included in Growth.',
		match: q(/\b(truck|freight|backhaul|return load|transport)/)
	},
	{
		code: 'starter.article9',
		tier: 'starter',
		target: 'growth',
		capability: 'cbam_article9_ccts',
		message: 'Article 9 / CCTS carbon-price deductions and full precursor-chain accounting are handled natively in Growth.',
		match: q(/(article\s*9|\bccts\b|carbon price|precursor)/)
	},
	{
		code: 'starter.revised_upload',
		tier: 'starter',
		target: 'growth',
		capability: 'tamper_audit_locker',
		message: 'Need tamper-evident change history and a dedicated CA verifier review workspace? Available in MSME Growth.',
		match: is('manual_effort', 'revised_upload')
	},
	{
		code: 'starter.capex',
		tier: 'starter',
		target: 'growth',
		capability: 'capex_roi_simulator',
		message: "Model solar capex, scrap-ratio shifts, and depreciation payback directly inside Growth's Decarbonisation Simulator.",
		match: q(/(solar|capex|scrap ratio|\broi\b)/)
	},
	{
		code: 'starter.whatsapp',
		tier: 'starter',
		target: 'growth',
		capability: 'whatsapp_floorbot',
		message: 'Collect shift production inputs directly from factory supervisors via WhatsApp vernacular floor-bot on Growth.',
		match: is('manual_effort', 'whatsapp_click')
	},
	{
		code: 'growth.supplier_network',
		limit: true,
		tier: 'growth',
		target: 'importer_hub',
		capability: 'importer_multi_supplier',
		message: 'Coordinating 15 to 500 independent supplier factories? EU Importer Hub provides isolated supplier portals and bulk invites.',
		match: (t) => (t.kind === 'limit_hit' && t.payload === 'add_plant' && (t.count ?? 0) >= 3) || (t.kind === 'entity_pattern' && t.payload === 'external_domain')
	},
	{
		code: 'growth.bulk_invite',
		tier: 'growth',
		target: 'importer_hub',
		capability: 'importer_multi_supplier',
		message: 'Automate bulk supplier invitation links, missing-evidence reminders, and submission tracking with EU Importer Hub.',
		match: (t) => (t.kind === 'entity_pattern' && t.payload === 'multi_supplier_csv') || q(/(bulk invite|supplier reminder)/)(t)
	},
	{
		code: 'growth.consolidated_xml',
		tier: 'growth',
		target: 'importer_hub',
		capability: 'consolidated_cbam_xml',
		message: 'Generate consolidated multi-supplier EU CBAM Registry XML packs and customs audit defence packs on EU Importer Hub.',
		match: is('entity_pattern', 'consolidated_xml')
	},
	{
		code: 'growth.fleet',
		tier: 'growth',
		target: 'importer_hub',
		capability: 'freightlink_enterprise_fleet',
		message: 'Multi-truck live fleet dashboards, bulk load matching, and cargo consolidation are part of EU Importer Hub.',
		match: (t) => (t.kind === 'entity_pattern' && t.payload === 'multi_load' && (t.count ?? 0) >= 3) || q(/(live fleet|multi-truck|multi truck|cargo consolidation)/)(t)
	},
	{
		code: 'growth.eu_port',
		tier: 'growth',
		target: 'importer_hub',
		capability: 'freightlink_enterprise_fleet',
		message: 'End-to-end international carrier/document coordination and green logistics scoring unlock with EU Importer Hub.',
		match: is('entity_pattern', 'eu_port')
	},
	{
		code: 'growth.benchmark',
		tier: 'growth',
		target: 'importer_hub',
		capability: 'supplier_benchmarking',
		message: 'Compare carbon intensity and CBAM tariff exposure across your entire supplier portfolio in EU Importer Hub.',
		match: q(/(supplier benchmark|tariff comparison)/)
	},
	{
		code: 'importer.extra_supplier',
		tier: 'importer_hub',
		target: 'expansion_addon',
		capability: 'extra_supplier_seat',
		message: 'Your 15 included supplier seats are active. Additional suppliers are billed transparently at $25/month each—confirm to invite.',
		match: is('limit_hit', 'invite_supplier', (t) => (t.count ?? 0) >= 15)
	}
];

export const HINT_COOLDOWN_MS = 14 * 24 * 60 * 60 * 1000;

/**
 * Returns null when the tenant is already entitled, the hint is cooling down, or nothing matches.
 * Invoked silently on Cmd+K keystrokes, form threshold hits and upload actions.
 */
export function evaluateSubtleHint(
	manifest: EntitlementManifest,
	trigger: HintTrigger,
	dismissals: Record<string, number> = {},
	now = Date.now()
): SubtleHint | null {
	for (const rule of HINT_RULES) {
		if (rule.tier !== manifest.tier) continue;
		if (!rule.match(trigger)) continue;
		// Capability-based rules are skipped if the tenant already holds the capability; limit-based rules
		// (plants, SKUs, supplier seats) are about quantity, so they bypass this check.
		if (!rule.limit && rule.capability !== 'extra_supplier_seat' && manifest.capabilities.has(rule.capability)) continue;
		const dismissedAt = dismissals[rule.code];
		if (dismissedAt && now - dismissedAt < HINT_COOLDOWN_MS) return null;
		return {
			code: rule.code,
			message: rule.message,
			suggested_tier: rule.target,
			capability: rule.capability,
			learn_more: `/settings/billing?focus=${rule.capability}`
		};
	}
	return null;
}
