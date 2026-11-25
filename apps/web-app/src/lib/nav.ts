/** Full navigation catalogue. AppShell filters it by manifest.visible_modules before rendering. */
export const NAV = [
	{ group: 'Overview', items: [{ key: 'dashboard', label: 'Dashboard', icon: 'dashboard' }, { key: 'inbox', label: 'Inbox', icon: 'inbox' }] },
	{
		group: 'Zevra Business',
		items: [
			{ key: 'plants', label: 'Plants & Zones', icon: 'plant' },
			{ key: 'products', label: 'Products & SKUs', icon: 'product' },
			{ key: 'inventory', label: 'Inventory', icon: 'inventory' },
			{ key: 'finance', label: 'Finance', icon: 'finance' },
			{ key: 'integrations/tally', label: 'Tally.ERP 9 Hub', icon: 'tally' }
		]
	},
	{
		group: 'Zevra Carbon',
		items: [
			{ key: 'carbon/inputs', label: 'Activity Inputs', icon: 'inputs' },
			{ key: 'carbon/cbam', label: 'CBAM Engine', icon: 'cbam' },
			{ key: 'carbon/exports', label: 'Registry Exports', icon: 'exports' },
			{ key: 'carbon/audit', label: 'Audit Locker', icon: 'audit' },
			{ key: 'carbon/simulator', label: 'Capex Simulator', icon: 'simulator' }
		]
	},
	{
		group: 'Trace & FreightLink',
		items: [
			{ key: 'trace/qr', label: 'QR Studio', icon: 'qr' },
			{ key: 'freight/marketplace', label: 'Backhaul Market', icon: 'route' },
			{ key: 'freight/shipments', label: 'Shipments', icon: 'truck' },
			{ key: 'freight/transporters', label: 'Transporters', icon: 'carrier' },
			{ key: 'freight/enterprise', label: 'Enterprise Fleet', icon: 'fleet' }
		]
	},
	{ group: 'Connect', items: [{ key: 'connect', label: 'Zevra Connect', icon: 'connect' }, { key: 'importer-hub', label: 'Importer Hub', icon: 'importer' }] },
	{ group: 'Settings', items: [{ key: 'settings/billing', label: 'Billing & Usage', icon: 'billing' }, { key: 'settings/team', label: 'Team & Access', icon: 'team' }] }
];

