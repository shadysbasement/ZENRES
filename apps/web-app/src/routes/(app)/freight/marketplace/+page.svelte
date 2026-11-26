<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Drawer from '$lib/components/ui/Drawer.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { rankMatches, rejected, createBooking, type Truck, type Load, type Match } from '$lib/freight';
	import { avoidedFuel } from '$lib/carbon';
	import { USER } from '$lib/demo/seed';

	const trucks = $derived(ws.freight.trucks.map((t) => ({ ...t, window: t.window as [string, string] })) as Truck[]);
	let loadIdx = $state(0);
	const load = $derived<Load>({ ...ws.freight.loads[loadIdx], window: ws.freight.loads[loadIdx].window as [string, string] });
	const matches = $derived(rankMatches(trucks, load));
	const rej = $derived(rejected(trucks, load));
	let tab = $state<'match' | 'capacity' | 'demand'>('match');
	let quote = $state<Match | null>(null);
	let accept = $state(false);
	const fee = 2;

	let demand = $state({ origin: 'Mumbai', dest: 'Delhi', weight: 10, cargo: 'Aluminium profiles', packaging: 'Crated', from: '2026-11-28T08:00', to: '2026-11-29T18:00', budget: 36000, port: '' });
	const EU_PORTS = ['Rotterdam', 'Antwerp', 'Hamburg', 'Genoa', 'Le Havre', 'Valencia'];
	let posted = $state(0);

	function postDemand(e: SubmitEvent) {
		e.preventDefault();
		if (demand.port && EU_PORTS.includes(demand.port)) {
			const h = ws.hint('freight', { kind: 'entity_pattern', payload: 'eu_port' });
			if (h) return;
		}
		ws.freight.loads.push({ id: `ld${Date.now()}`, shipper: 'Deccan Alloy Works', origin: demand.origin, dest: demand.dest, corridor: demand.origin === 'Mumbai' ? 'BOM-DEL' : 'DEL-LDH', weight: demand.weight, cargo: demand.cargo, packaging: demand.packaging, window: [demand.from, demand.to], budget_inr: demand.budget });
		posted++;
		ws.hint('freight', { kind: 'entity_pattern', payload: 'multi_load', count: posted });
		ws.toast('Load demand posted to the corridor');
	}
	function book() {
		if (!quote) return;
		const b = createBooking(quote, { accepted: accept, authorisedBy: USER.name, feePct: fee });
		const fuel = avoidedFuel(410, 322);
		ws.freight.shipments.unshift({ id: `sh${Date.now()}`, ref: `FL-2611-0${20 + ws.freight.shipments.length}`, route: `${ws.freight.loads[loadIdx].origin} → ${ws.freight.loads[loadIdx].dest}`, corridor: load.corridor, carrier: quote.truck.carrier, vehicle: quote.truck.reg, status: 'confirmed', gps_consent: false, freight_inr: b.freight_inr, fee_pct: fee, baseline_l: 410, matched_l: 322, pickup_scanned: false, delivery_scanned: false, progress: 0, eta: load.window[1], container: null });
		ws.toast(`Booked ${quote.truck.reg} · ₹${b.total_inr.toLocaleString('en-IN')} incl. disclosed fee · saves ~${fuel.litres} L`);
		quote = null;
		accept = false;
	}
	const reasonLabel: Record<string, string> = { corridor: 'Different corridor', window: 'Window mismatch', payload: 'Insufficient payload', vehicle_type: 'Vehicle type', documents: 'Documents invalid' };
</script>

<svelte:head><title>FreightLink Marketplace — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Trace & FreightLink" title="Backhaul" serif="marketplace" lede="Empty-truck finder, return-capacity and load-demand posting, and a deterministic corridor matcher. Nothing is ever auto-booked." />

<div class="tabs-line" role="tablist">
	<button role="tab" aria-selected={tab === 'match'} onclick={() => (tab = 'match')}>Corridor matcher</button>
	<button role="tab" aria-selected={tab === 'capacity'} onclick={() => (tab = 'capacity')}>Empty truck finder</button>
	<button role="tab" aria-selected={tab === 'demand'} onclick={() => (tab = 'demand')}>Post load demand</button>
</div>
<HintWhisper context="freight" />

{#if tab === 'match'}
	<div class="g2">
		<section class="panel">
			<div class="panel-head"><h2>Load demand</h2></div>
			{#each ws.freight.loads as l, i (l.id)}
				<button class="load" class:on={loadIdx === i} onclick={() => (loadIdx = i)}>
					<span class="mono corr">{l.corridor}</span>
					<strong>{l.origin} → {l.dest}</strong>
					<span class="small muted">{l.weight} t · {l.cargo} · budget ₹{l.budget_inr.toLocaleString('en-IN')}</span>
				</button>
			{/each}
			<div class="filters">
				<span class="eyebrow">Hard filters</span>
				{#each ['Corridor', 'Departure window', 'Payload ≥ weight', 'Vehicle & cargo rules', 'Valid RC / insurance / permit'] as f (f)}<span class="pill pill-outline">✓ {f}</span>{/each}
				<span class="eyebrow" style="margin-top:8px">Ranking</span>
				<span class="mono small">0.35 detour + 0.30 quote + 0.25 reliability + 0.10 availability</span>
				<span class="small muted">Paid promotion is not a ranking input.</span>
			</div>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Ranked matches</h2><span class="mono small muted">{matches.length} eligible · {rej.length} filtered</span></div>
			{#each matches as m, i (m.truck.id)}
				<div class="match" style="--i:{i}">
					<span class="rank mono">#{i + 1}</span>
					<div class="mi">
						<div class="row"><strong class="mono">{m.truck.reg}</strong>{#if m.truck.verified}<span class="pill pill-emerald">Verified</span>{/if}</div>
						<span class="small">{m.truck.carrier} · {m.truck.vehicle} · {m.truck.payload} t</span>
						<span class="mono small muted">{m.reasons.join(' · ')}</span>
					</div>
					<div class="score"><span class="mono">{(m.score * 100).toFixed(0)}</span><div class="bar"><span style="width:{m.score * 100}%"></span></div></div>
					<button class="btn btn-sm btn-dark" onclick={() => (quote = m)}>Review quote</button>
				</div>
			{:else}<p class="empty">No eligible return capacity on this corridor yet.</p>{/each}
			{#if rej.length}
				<div class="rej">
					<span class="eyebrow">Filtered out</span>
					{#each rej as r (r.truck.id)}<div class="small"><span class="mono">{r.truck.reg}</span> — {reasonLabel[r.reason!]}</div>{/each}
				</div>
			{/if}
		</section>
	</div>
{:else if tab === 'capacity'}
	<section class="panel">
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>Vehicle</th><th>Carrier</th><th>Type</th><th class="num">Payload</th><th>Corridor</th><th>Window</th><th class="num">Quote</th><th>Status</th></tr></thead>
				<tbody>{#each ws.freight.trucks as t (t.id)}<tr><td class="mono strong">{t.reg}</td><td>{t.carrier}</td><td>{t.vehicle}</td><td class="num">{t.payload} t</td><td class="mono">{t.origin} → {t.dest}</td><td class="small">{t.window[0].replace('T', ' ')} → {t.window[1].slice(5).replace('T', ' ')}</td><td class="num">₹{t.quote_inr.toLocaleString('en-IN')}</td><td>{#if t.verified}<span class="pill pill-emerald">Verified</span>{:else}<span class="pill pill-terracotta">Docs pending</span>{/if}</td></tr>{/each}</tbody>
			</table>
		</div>
		<p class="fine">Stale capacity listings auto-expire at the end of their departure window.</p>
	</section>
{:else}
	<form class="panel demand" onsubmit={postDemand}>
		<div class="grid3">
			<label class="field"><span class="label">Pickup</span><input class="input" bind:value={demand.origin} /></label>
			<label class="field"><span class="label">Drop</span><input class="input" bind:value={demand.dest} /></label>
			<label class="field"><span class="label">Weight (t)</span><input class="input num" type="number" bind:value={demand.weight} /></label>
			<label class="field"><span class="label">Cargo type</span><input class="input" bind:value={demand.cargo} /></label>
			<label class="field"><span class="label">Packaging</span><select class="select" bind:value={demand.packaging}><option>Crated</option><option>Palletised</option><option>Coils</option><option>Loose</option></select></label>
			<label class="field"><span class="label">Budget (₹)</span><input class="input num" type="number" bind:value={demand.budget} /></label>
			<label class="field"><span class="label">Earliest pickup</span><input class="input" type="datetime-local" bind:value={demand.from} /></label>
			<label class="field"><span class="label">Latest delivery</span><input class="input" type="datetime-local" bind:value={demand.to} /></label>
			<label class="field"><span class="label">Onward export port <span class="muted">(optional)</span></span><select class="select" bind:value={demand.port}><option value="">—</option><option>JNPT Nhava Sheva</option><option>Mundra</option>{#each EU_PORTS as p (p)}<option>{p}</option>{/each}</select></label>
		</div>
		<button class="btn btn-primary"><Icon name="plus" size={14} />Post load demand</button>
	</form>
{/if}

<Drawer open={!!quote} title="Price negotiation" eyebrow={quote ? `${quote.truck.reg} · ${quote.truck.carrier}` : ''} onclose={() => (quote = null)}>
	{#if quote}
		<div class="q">
			<div><span>Line-haul freight</span><span class="num">₹{(quote.truck.quote_inr * 0.86).toLocaleString('en-IN', { maximumFractionDigits: 0 })}</span></div>
			<div><span>Loading / unloading</span><span class="num">₹{(quote.truck.quote_inr * 0.04).toLocaleString('en-IN', { maximumFractionDigits: 0 })}</span></div>
			<div><span>Tolls (FASTag est.)</span><span class="num">₹{(quote.truck.quote_inr * 0.1).toLocaleString('en-IN', { maximumFractionDigits: 0 })}</span></div>
			<div class="t"><span>Freight (GST extra @ 12% RCM)</span><span class="num">₹{quote.truck.quote_inr.toLocaleString('en-IN')}</span></div>
			<div class="fee"><span>FreightLink fee (disclosed, illustrative {fee}%)</span><span class="num">₹{(quote.truck.quote_inr * fee / 100).toLocaleString('en-IN')}</span></div>
			<p class="small muted">Cancellation: free until 6 h before pickup, then 15% of freight. Detention ₹1,500 per 12 h after 4 free hours.</p>
		</div>
		<label class="accept"><input type="checkbox" bind:checked={accept} /> I, {USER.name}, am authorised to accept this itemised quote on behalf of Deccan Alloy Works.</label>
		<button class="btn btn-sm btn-ghost" onclick={() => ws.toast('Counter-offer sent: ₹36,000')}>Counter-offer ₹36,000</button>
	{/if}
	{#snippet footer()}
		<button class="btn btn-sm btn-ghost" onclick={() => (quote = null)}>Cancel</button>
		<button class="btn btn-sm btn-primary" disabled={!accept} onclick={book}>Confirm booking</button>
	{/snippet}
</Drawer>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1fr 1.6fr;
		gap: 16px;
		align-items: start;
	}
	.small {
		font-size: 11.5px;
	}
	.strong {
		color: var(--ink-primary);
	}
	.load {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 3px;
		padding: 12px;
		margin-bottom: 8px;
		text-align: left;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		font-size: 13px;
		transition: all 200ms;
	}
	.load.on {
		border-color: var(--accent-cobalt);
		box-shadow: 0 0 0 3px var(--accent-cobalt-tint);
	}
	.corr {
		font-size: 10.5px;
		color: var(--accent-cobalt);
	}
	.filters {
		display: flex;
		flex-direction: column;
		gap: 6px;
		align-items: flex-start;
		margin-top: 14px;
		padding-top: 14px;
		border-top: var(--hairline);
	}
	.match {
		display: grid;
		grid-template-columns: 36px 1fr 110px auto;
		gap: 12px;
		align-items: center;
		padding: 14px 0;
		border-bottom: var(--hairline);
		animation: zv-fade-in 500ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 80ms);
	}
	.rank {
		font-size: 1.1rem;
		color: var(--brand-emerald);
	}
	.mi {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}
	.row {
		display: flex;
		gap: 8px;
		align-items: center;
	}
	.score {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12px;
	}
	.rej {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		color: var(--ink-muted);
	}
	.demand {
		display: flex;
		flex-direction: column;
		gap: 16px;
		align-items: flex-start;
	}
	.grid3 {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
		width: 100%;
	}
	.q > div {
		display: flex;
		justify-content: space-between;
		padding: 10px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
		color: var(--ink-primary);
	}
	.q .t {
		font-weight: 600;
	}
	.q .fee {
		color: var(--accent-cobalt);
	}
	.q p {
		margin: 10px 0;
	}
	.accept {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		padding: 12px;
		margin: 10px 0;
		border-radius: var(--radius-sm);
		background: var(--surface-subtle);
		font-size: 12.5px;
		accent-color: var(--brand-emerald);
	}
	@media (max-width: 1100px) {
		.g2,
		.grid3 {
			grid-template-columns: 1fr;
		}
	}
</style>
