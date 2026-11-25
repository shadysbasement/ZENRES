<script lang="ts">
	import { page } from '$app/state';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { TIER_LABEL, type SubscriptionTier } from '$lib/entitlements/manifest';
	import { money } from '$lib/format';

	const focus = page.url.searchParams.get('focus');
	let cycle = $state<'monthly' | 'annual'>('monthly');
	const PLANS: { id: SubscriptionTier; m: number; y: number; caps: string[] }[] = [
		{ id: 'starter', m: 99, y: 990, caps: ['core_erp_tally_csv', 'basic_cbam', 'evidence_vault', 'qr_traceability', 'zevra_connect'] },
		{ id: 'growth', m: 249, y: 2490, caps: ['ocr_ingestion', 'whatsapp_floorbot', 'precursor_tracking', 'cbam_article9_ccts', 'tamper_audit_locker', 'capex_roi_simulator', 'priority_ca_workspace', 'freightlink_core'] },
		{ id: 'importer_hub', m: 499, y: 4990, caps: ['importer_multi_supplier', 'consolidated_cbam_xml', 'supplier_benchmarking', 'customs_audit_pack', 'freightlink_enterprise_fleet', 'ai_backhaul_bulk', 'extra_supplier_seat'] }
	];
	const CAP_LABEL: Record<string, string> = {
		core_erp_tally_csv: 'Tally CSV & business core', basic_cbam: 'Basic Scope 1 & 2 CBAM', evidence_vault: 'Evidence vault', qr_traceability: 'Dual QR traceability', zevra_connect: 'Zevra Connect',
		ocr_ingestion: 'OCR bill extraction', whatsapp_floorbot: 'WhatsApp floor-bot', precursor_tracking: 'Unlimited SKUs & precursors', cbam_article9_ccts: 'Article 9 / CCTS', tamper_audit_locker: 'Tamper-evident audit locker', capex_roi_simulator: 'Capex ROI simulator', priority_ca_workspace: 'Priority CA workspace', freightlink_core: 'FreightLink backhaul',
		importer_multi_supplier: '15–500 supplier portal', consolidated_cbam_xml: 'Consolidated Registry XML', supplier_benchmarking: 'Supplier benchmarking', customs_audit_pack: 'Customs audit defence pack', freightlink_enterprise_fleet: 'Enterprise fleet', ai_backhaul_bulk: 'AI / bulk backhaul', extra_supplier_seat: 'Extra supplier seats $25/mo'
	};
	const quotas = [
		['OCR pages', 412, 2000],
		['GPS / map API calls', 3180, 10000],
		['Evidence storage (GB)', 3.4, 25],
		['Seats', 7, 15]
	] as const;
	const ledger = [
		['FL-2611-014', 'Chakan → JNPT', 21400],
		['FL-2611-011', 'Delhi → Ludhiana', 14800]
	] as const;
</script>

<svelte:head><title>Billing & Usage — Zevra</title></svelte:head>

<PageHeader eyebrow="Settings · E1" title="Subscription, billing &" serif="usage" lede="Transparent plan management. Prices in USD excluding applicable taxes; INR and EUR shown for convenience.">
	{#snippet actions()}
		<div class="cyc">
			<button class:on={cycle === 'monthly'} onclick={() => (cycle = 'monthly')}>Monthly</button>
			<button class:on={cycle === 'annual'} onclick={() => (cycle = 'annual')}>Annual · 2 months free</button>
		</div>
	{/snippet}
</PageHeader>

<div class="plans">
	{#each PLANS as p, i (p.id)}
		{@const current = ws.tier === p.id}
		<article class="panel plan fade-in" class:current style="--i:{i}">
			<div class="row"><span class="eyebrow">{TIER_LABEL[p.id]}</span>{#if current}<span class="pill pill-emerald">Current plan</span>{/if}</div>
			<div class="price"><span class="num">{money(cycle === 'monthly' ? p.m : p.y, ws.currency)}</span><span class="muted mono">/ {cycle === 'monthly' ? 'mo' : 'yr'}</span></div>
			<ul role="list">{#each p.caps as c (c)}<li class:focus={focus === c}>{CAP_LABEL[c]}</li>{/each}</ul>
			{#if !current}<button class="btn btn-sm {i > PLANS.findIndex((x) => x.id === ws.tier) ? 'btn-primary' : ''}" onclick={() => { ws.setTier(p.id); ws.toast(`Workspace moved to ${TIER_LABEL[p.id]}`); }}>{i > PLANS.findIndex((x) => x.id === ws.tier) ? 'Upgrade' : 'Switch'} to {TIER_LABEL[p.id]}</button>{/if}
		</article>
	{/each}
</div>

<div class="g3">
	<section class="panel">
		<div class="panel-head"><h2>Fair-use quotas</h2></div>
		{#each quotas as [l, used, cap] (l)}
			<div class="q"><div class="row"><span>{l}</span><span class="mono small">{used} / {cap}</span></div><div class="bar"><span style="width:{(used / cap) * 100}%"></span></div></div>
		{/each}
	</section>
	<section class="panel">
		<div class="panel-head"><h2>Add-ons</h2></div>
		<div class="addon"><div><strong>Assisted setup sprint</strong><p class="small muted">Ledger, meter and CN mapping by our team.</p></div><span class="num">{money(350, ws.currency)}</span></div>
		<div class="addon"><div><strong>Extra supplier seats</strong><p class="small muted">Beyond 15 included · always confirmed before invite.</p></div><span class="num">{money(25, ws.currency)}/mo</span></div>
	</section>
	<section class="panel">
		<div class="panel-head"><h2>FreightLink commission ledger</h2></div>
		{#each ledger as [ref, route, amt] (ref)}<div class="addon"><div><span class="mono">{ref}</span><p class="small muted">{route} · disclosed before booking</p></div><span class="num">₹{(amt * 0.02).toLocaleString('en-IN')}</span></div>{/each}
		<p class="fine">Illustrative 2% commission. Freight is paid separately to the carrier.</p>
	</section>
</div>

<style>
	.cyc {
		display: flex;
		padding: 3px;
		border-radius: var(--radius-sm);
		background: var(--surface-subtle);
		border: var(--hairline);
	}
	.cyc button {
		height: 28px;
		padding: 0 12px;
		border-radius: 3px;
		font-size: 12px;
		color: var(--ink-muted);
	}
	.cyc button.on {
		background: var(--surface-card);
		color: var(--ink-primary);
		box-shadow: 0 0 0 1px var(--border-hairline);
	}
	.plans {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		margin-bottom: 16px;
	}
	.plan {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.plan.current {
		border-color: var(--brand-emerald);
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
	}
	.row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
	}
	.price .num {
		font-size: 2rem;
		letter-spacing: -0.04em;
		color: var(--ink-primary);
	}
	.plan ul {
		display: flex;
		flex-direction: column;
		gap: 6px;
		flex: 1;
	}
	.plan li {
		font-size: 12.5px;
		padding: 4px 6px;
		border-radius: 3px;
	}
	.plan li.focus {
		background: var(--brand-emerald-tint);
		color: var(--brand-emerald);
		font-weight: 600;
		animation: pulse 1.6s ease-in-out 2;
	}
	@keyframes pulse {
		50% { box-shadow: 0 0 0 4px var(--brand-emerald-tint); }
	}
	.g3 {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
	}
	.q {
		margin-bottom: 12px;
		font-size: 12.5px;
	}
	.q .bar {
		margin-top: 6px;
	}
	.small {
		font-size: 11.5px;
	}
	.addon {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
	}
	@media (max-width: 1100px) {
		.plans,
		.g3 {
			grid-template-columns: 1fr;
		}
	}
</style>
