<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { avoidedFuel } from '$lib/carbon';
	import { fmtDate } from '$lib/format';

	const STAGES = ['confirmed', 'picked_up', 'in_transit', 'delivered'] as const;
	let sel = $state(ws.freight.shipments[0]?.id);
	const s = $derived(ws.freight.shipments.find((x) => x.id === sel) ?? ws.freight.shipments[0]);
	const fuel = $derived(avoidedFuel(s.baseline_l, s.matched_l));
	const totals = $derived(ws.freight.shipments.reduce((a, x) => { const f = avoidedFuel(x.baseline_l, x.matched_l); return { l: a.l + f.litres, kg: a.kg + f.kgCO2, inr: a.inr + x.freight_inr }; }, { l: 0, kg: 0, inr: 0 }));
	let scanning = $state<'pickup' | 'delivery' | null>(null);
	let dispute = $state(false);
	let podName = $state('');

	function scan(kind: 'pickup' | 'delivery') {
		scanning = kind;
		setTimeout(() => {
			if (kind === 'pickup') {
				s.pickup_scanned = true;
				s.status = 'in_transit';
				s.progress = 0.08;
			} else {
				s.delivery_scanned = true;
				s.status = 'delivered';
				s.progress = 1;
			}
			scanning = null;
			ws.toast(`${kind === 'pickup' ? 'Pickup' : 'Delivery'} QR verified for ${s.ref}`);
		}, 1400);
	}
	const docs = ['Freight invoice', 'Delivery challan', 'E-way bill', 'Lorry receipt (LR)', 'Shipping bill (export)'];
	const option = $derived({
		tooltip: { trigger: 'axis' },
		legend: { top: 0, right: 0 },
		grid: { top: 30, left: 4, right: 4, bottom: 0, containLabel: true },
		xAxis: { type: 'category', data: ws.freight.shipments.map((x) => x.ref.slice(-3)) },
		yAxis: { type: 'value', name: 'L', nameTextStyle: { fontSize: 10 } },
		series: [
			{ name: 'Baseline (empty return + separate trip)', type: 'bar', barWidth: 12, data: ws.freight.shipments.map((x) => x.baseline_l), itemStyle: { color: '#C5CFCA', borderRadius: [2, 2, 0, 0] } },
			{ name: 'Matched journey', type: 'bar', barWidth: 12, data: ws.freight.shipments.map((x) => x.matched_l), itemStyle: { color: '#1E4D6B', borderRadius: [2, 2, 0, 0] } }
		]
	});
</script>

<svelte:head><title>Shipments — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Trace & FreightLink · C3" title="Shipments, tracking &" serif="proof of delivery" lede="Booked → Pickup → In transit → Delivered, with QR scans at both ends, consent-based GPS, a digital POD and a document vault." />

<div class="list">
	{#each ws.freight.shipments as x (x.id)}
		<button class="ship" class:on={x.id === s.id} onclick={() => (sel = x.id)}>
			<span class="mono">{x.ref}</span><strong>{x.route}</strong>
			<span class="pill" class:pill-emerald={x.status === 'delivered'} class:pill-cobalt={x.status === 'in_transit'}>{x.status.replace('_', ' ')}</span>
		</button>
	{/each}
</div>

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>{s.ref} · {s.route}</h2><span class="mono small muted">{s.carrier} · {s.vehicle}</span></div>
		<ol class="life" role="list">
			{#each STAGES as st, i (st)}
				{@const idx = STAGES.indexOf(s.status as (typeof STAGES)[number])}
				<li class:done={i <= idx}><span class="n"></span><span>{st.replace('_', ' ')}</span></li>
			{/each}
		</ol>
		<div class="track">
			<div class="road"><span style="width:{s.progress * 100}%"></span><i style="left:{s.progress * 100}%"><Icon name="truck" size={16} /></i></div>
			<div class="ends mono small"><span>{s.route.split(' → ')[0]}</span><span>ETA {fmtDate(s.eta)}</span><span>{s.route.split(' → ')[1]}</span></div>
		</div>
		<div class="row">
			<label class="gps"><input type="checkbox" bind:checked={s.gps_consent} /> Carrier GPS consent {s.gps_consent ? '(granted for this trip)' : '(not granted — milestones only)'}</label>
		</div>
		<div class="scans">
			<button class="btn" disabled={s.pickup_scanned || !!scanning} onclick={() => scan('pickup')}><Icon name="qr" size={15} />{s.pickup_scanned ? 'Pickup QR ✓' : scanning === 'pickup' ? 'Scanning…' : 'Scan pickup QR'}</button>
			<button class="btn" disabled={!s.pickup_scanned || s.delivery_scanned || !!scanning} onclick={() => scan('delivery')}><Icon name="qr" size={15} />{s.delivery_scanned ? 'Delivery QR ✓' : scanning === 'delivery' ? 'Scanning…' : 'Scan delivery QR'}</button>
		</div>
		{#if scanning}<div class="scanner"><span></span></div>{/if}
		{#if s.status === 'delivered'}
			<div class="pod card">
				<span class="eyebrow">Digital POD</span>
				<label class="field"><span class="label">Received by</span><input class="input" bind:value={podName} placeholder="Consignee name" /></label>
				<button class="btn btn-sm btn-dark" disabled={!podName} onclick={() => ws.toast(`POD signed by ${podName} · freight invoice queued for Tally`)}>Sign POD & sync invoice to Tally</button>
			</div>
		{/if}
		<div class="docs">
			<span class="eyebrow">Transport document vault</span>
			<div class="dl">{#each docs as d, i (d)}<span class="pill" class:pill-emerald={i < 3}>{i < 3 ? '✓ ' : ''}{d}</span>{/each}</div>
		</div>
		<button class="btn btn-sm btn-ghost" onclick={() => (dispute = !dispute)}><Icon name="alert" size={14} />Raise dispute / claim</button>
		{#if dispute}
			<div class="disp card">
				<select class="select"><option>Short delivery</option><option>Damage in transit</option><option>Detention charges</option><option>Delay</option></select>
				<textarea class="textarea" placeholder="Describe the issue and attach photos…"></textarea>
				<button class="btn btn-sm btn-dark" onclick={() => { dispute = false; ws.toast('Dispute escalated to FreightLink trust team'); }}>Escalate</button>
			</div>
		{/if}
	</section>
	<div class="side">
		<section class="panel">
			<div class="panel-head"><h2>Avoided fuel — this trip</h2><span class="pill pill-cobalt">Logistics estimate</span></div>
			<div class="fuel"><span class="num big">{fuel.litres} L</span><span class="muted">≈ {fuel.kgCO2} kg CO₂</span></div>
			<p class="mono small">{s.baseline_l} L baseline → {s.matched_l} L matched</p>
			<p class="fine">Reported separately from CBAM embedded emissions and guarded against double counting.</p>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Freight cost & avoided emissions</h2></div>
			<EChart {option} height={180} />
			<div class="tot mono small"><span>Spend ₹{totals.inr.toLocaleString('en-IN')}</span><span>Saved {totals.l} L · {(totals.kg / 1000).toFixed(2)} t CO₂</span></div>
		</section>
	</div>
</div>

<style>
	.list {
		display: flex;
		gap: 8px;
		margin-bottom: 16px;
		overflow-x: auto;
	}
	.ship {
		display: flex;
		flex-direction: column;
		gap: 4px;
		align-items: flex-start;
		padding: 12px 14px;
		min-width: 220px;
		border: var(--hairline);
		border-radius: var(--radius-md);
		background: var(--surface-card);
		font-size: 13px;
		text-align: left;
	}
	.ship.on {
		border-color: var(--obsidian);
		box-shadow: 0 0 0 1px var(--obsidian);
	}
	.ship .mono {
		font-size: 10.5px;
		color: var(--ink-muted);
	}
	.g2 {
		display: grid;
		grid-template-columns: 1.5fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.small {
		font-size: 11.5px;
	}
	.life {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		margin-bottom: 20px;
	}
	.life li {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 12px;
		text-transform: capitalize;
		color: var(--ink-muted);
		border-top: 2px solid var(--surface-sunken);
		padding-top: 10px;
	}
	.life li.done {
		color: var(--ink-primary);
		border-color: var(--brand-emerald);
	}
	.track {
		padding: 16px 0;
	}
	.road {
		position: relative;
		height: 4px;
		background: repeating-linear-gradient(90deg, var(--border-hairline) 0 8px, transparent 8px 14px);
		border-radius: 2px;
	}
	.road span {
		position: absolute;
		inset: 0 auto 0 0;
		background: var(--accent-cobalt);
		border-radius: 2px;
		transition: width 1.2s var(--ease-expo);
	}
	.road i {
		position: absolute;
		top: -13px;
		transform: translateX(-50%);
		width: 30px;
		height: 30px;
		display: grid;
		place-items: center;
		border-radius: 50%;
		background: var(--surface-card);
		border: 1.5px solid var(--accent-cobalt);
		color: var(--accent-cobalt);
		transition: left 1.2s var(--ease-expo);
	}
	.ends {
		display: flex;
		justify-content: space-between;
		margin-top: 18px;
		color: var(--ink-muted);
	}
	.gps {
		display: flex;
		gap: 8px;
		font-size: 12.5px;
		accent-color: var(--brand-emerald);
	}
	.scans {
		display: flex;
		gap: 8px;
		margin: 14px 0;
	}
	.scanner {
		position: relative;
		height: 120px;
		border-radius: var(--radius-md);
		background: var(--obsidian);
		overflow: hidden;
		margin-bottom: 14px;
	}
	.scanner span {
		position: absolute;
		left: 10%;
		right: 10%;
		height: 2px;
		background: var(--vault-phosphor);
		box-shadow: 0 0 16px var(--vault-phosphor);
		animation: scan 1.2s ease-in-out infinite alternate;
	}
	@keyframes scan {
		from { top: 15%; }
		to { top: 85%; }
	}
	.pod,
	.disp {
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-bottom: 14px;
	}
	.docs {
		margin: 6px 0 12px;
	}
	.dl {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: 8px;
	}
	.fuel {
		display: flex;
		align-items: baseline;
		gap: 10px;
	}
	.big {
		font-size: 2rem;
		letter-spacing: -0.04em;
		color: var(--accent-cobalt);
	}
	.tot {
		display: flex;
		justify-content: space-between;
		margin-top: 8px;
		color: var(--ink-muted);
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
	}
</style>
