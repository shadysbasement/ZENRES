/**
 * Silent Tier Gating — entitlement manifest contract (mirrors services/core-api/internal/entitlements).
 * Navigation is built ONLY from `visible_modules`; unentitled modules never reach the DOM.
 */
export type SubscriptionTier = 'starter' | 'growth' | 'importer_hub';

export type Capability =
	| 'core_erp_tally_csv'
	| 'basic_cbam'
	| 'evidence_vault'
	| 'qr_traceability'
	| 'zevra_connect'
	| 'ocr_ingestion'
	| 'whatsapp_floorbot'
	| 'precursor_tracking'
	| 'cbam_article9_ccts'
	| 'tamper_audit_locker'
	| 'capex_roi_simulator'
	| 'priority_ca_workspace'
	| 'freightlink_core'
	| 'importer_multi_supplier'
	| 'consolidated_cbam_xml'
	| 'supplier_benchmarking'
	| 'customs_audit_pack'
	| 'freightlink_enterprise_fleet'
	| 'ai_backhaul_bulk';

export interface EntitlementManifest {
	tier: SubscriptionTier;
	limits: {
		max_plants: number | null;
		max_cn_skus: number | null;
		included_suppliers: number;
		extra_supplier_usd: number;
	};
	capabilities: Set<Capability>;
	visible_modules: string[];
}

const STARTER: Capability[] = ['core_erp_tally_csv', 'basic_cbam', 'evidence_vault', 'qr_traceability', 'zevra_connect'];
const GROWTH: Capability[] = [
	...STARTER,
	'ocr_ingestion',
	'whatsapp_floorbot',
	'precursor_tracking',
	'cbam_article9_ccts',
	'tamper_audit_locker',
	'capex_roi_simulator',
	'priority_ca_workspace',
	'freightlink_core'
];
const IMPORTER: Capability[] = [
	...GROWTH,
	'importer_multi_supplier',
	'consolidated_cbam_xml',
	'supplier_benchmarking',
	'customs_audit_pack',
	'freightlink_enterprise_fleet',
	'ai_backhaul_bulk'
];

/** module key → capability required (undefined = all tiers) */
export const MODULE_GATES: Record<string, Capability | undefined> = {
	dashboard: undefined,
	plants: undefined,
	products: undefined,
	inventory: undefined,
	finance: undefined,
	inbox: undefined,
	'integrations/tally': undefined,
	'carbon/inputs': undefined,
	'carbon/cbam': undefined,
	'carbon/exports': undefined,
	'carbon/audit': undefined,
	'carbon/simulator': 'capex_roi_simulator',
	'trace/qr': undefined,
	'freight/marketplace': 'freightlink_core',
	'freight/shipments': 'freightlink_core',
	'freight/transporters': 'freightlink_core',
	'freight/enterprise': 'freightlink_enterprise_fleet',
	connect: undefined,
	'importer-hub': 'importer_multi_supplier',
	'settings/billing': undefined,
	'settings/team': undefined
};

export function buildManifest(tier: SubscriptionTier): EntitlementManifest {
	const caps = new Set<Capability>(tier === 'starter' ? STARTER : tier === 'growth' ? GROWTH : IMPORTER);
	return {
		tier,
		limits: {
			max_plants: tier === 'starter' ? 1 : tier === 'growth' ? 3 : null,
			max_cn_skus: tier === 'starter' ? 5 : null,
			included_suppliers: tier === 'importer_hub' ? 15 : 0,
			extra_supplier_usd: 25
		},
		capabilities: caps,
		visible_modules: Object.entries(MODULE_GATES)
			.filter(([, cap]) => !cap || caps.has(cap))
			.map(([m]) => m)
	};
}

export const TIER_LABEL: Record<SubscriptionTier, string> = {
	starter: 'MSME Starter',
	growth: 'MSME Growth',
	importer_hub: 'EU Importer Hub'
};

/** Resolve a pathname like /freight/marketplace/123 to its module key. */
export function moduleForPath(pathname: string): string | null {
	const clean = pathname.replace(/^\/+|\/+$/g, '');
	if (!clean) return 'dashboard';
	const keys = Object.keys(MODULE_GATES).sort((a, b) => b.length - a.length);
	return keys.find((k) => clean === k || clean.startsWith(k + '/')) ?? null;
}
