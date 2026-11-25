<script lang="ts">
	import Drawer from '../ui/Drawer.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { QUALITY_LABEL, QUALITY_PILL, num } from '$lib/format';

	const zone = $derived(ws.zones.find((z) => z.code === ws.selectedZone) ?? null);
	const batches = $derived(zone ? ws.batches.filter((b) => b.zone_code === zone.code) : []);
	const product = (id: string) => ws.products.find((p) => p.id === id);
	const tco2hr = $derived(zone ? zone.live_power_kw * 0.000727 + zone.fuel_lph * 0.00268 : 0);
	let headcount = $state(0);
	$effect(() => {
		if (zone) headcount = zone.active_workers;
	});

	function saveHeadcount() {
		if (!zone) return;
		ws.updateZone(zone.id, { active_workers: Math.max(0, headcount) });
		ws.toast(`${zone.code} headcount updated to ${headcount} — broadcast to all open tabs`);
	}
	const lowStock = $derived(ws.inventory.filter((i) => i.qty < i.reorder));
</script>

<Drawer open={!!zone} eyebrow={zone ? `Zone inspector · ${zone.code}` : ''} title={zone?.name ?? ''} width={460} onclose={() => (ws.selectedZone = null)}>
	{#if zone}
		<div class="stats">
			<div><span class="eyebrow">On shift</span><span class="num big">{zone.active_workers}<small>/{zone.shift_target_workers}</small></span></div>
			<div><span class="eyebrow">Supervisor</span><span>{zone.supervisor}</span></div>
			<div><span class="eyebrow">Live power</span><span class="num">{num(zone.live_power_kw)} kW</span></div>
			<div><span class="eyebrow">Est. emissions</span><span class="num">{tco2hr.toFixed(3)} tCO₂e/h</span></div>
		</div>

		<div class="roster">
			{#each Array(zone.shift_target_workers) as _, i (i)}
				<span class:on={i < zone.active_workers} style="--i:{i}"></span>
			{/each}
		</div>

		<div class="ctl card">
			<span class="label">Update shift headcount</span>
			<div class="stepper">
				<button class="btn btn-sm" onclick={() => (headcount = Math.max(0, headcount - 1))}>−</button>
				<input class="input num" type="number" bind:value={headcount} min="0" max="99" />
				<button class="btn btn-sm" onclick={() => (headcount += 1)}>+</button>
				<button class="btn btn-sm btn-primary" onclick={saveHeadcount} disabled={headcount === zone.active_workers}>Save</button>
			</div>
		</div>

		{#if zone.type === 'office' && zone.staff}
			<h3 class="sec">Staff present</h3>
			<div class="staff">
				{#each Object.entries(zone.staff) as [k, v] (k)}
					<div><span class="num">{v}</span><span>{k}</span></div>
				{/each}
			</div>
			<ul class="list" role="list">
				<li><span class="dot text-emerald"></span>Tally.ERP 9 last sync 14 min ago · 2 vouchers pending review</li>
				<li><span class="dot text-terracotta"></span>October MSEDCL bill missing — Scope 2 estimated</li>
				<li><span class="dot text-cobalt"></span>2 open QA checks · 1 EU buyer RFQ awaiting quote</li>
			</ul>
		{:else if zone.type === 'warehouse'}
			<h3 class="sec">Stock vs reorder threshold</h3>
			{#each ws.inventory.filter((i) => i.category === 'raw_material' || i.category === 'finished_good') as it (it.id)}
				<div class="stock">
					<div class="row"><span>{it.name}</span><span class="num" class:text-terracotta={it.qty < it.reorder}>{num(it.qty, 1)} {it.unit}</span></div>
					<div class="bar"><span style="width:{Math.min(100, (it.qty / (it.reorder * 2.2)) * 100)}%;{it.qty < it.reorder ? 'background:var(--status-terracotta)' : ''}"></span></div>
					{#if it.lot}<span class="lot mono">lot {it.lot}</span>{/if}
				</div>
			{/each}
			{#if lowStock.length}<p class="warn">{lowStock.length} items below reorder level.</p>{/if}
		{:else}
			<h3 class="sec">Active SKUs & batches</h3>
			{#each batches as b (b.id)}
				{@const p = product(b.product_id)}
				{@const prog = zone.type === 'packaging' ? b.packaged_qty / b.target_qty : b.produced_qty / b.target_qty}
				<div class="batch">
					<div class="row">
						<span><span class="mono strong">{b.code}</span> · {p?.name}</span>
						<span class="pill {QUALITY_PILL[b.quality]}">{QUALITY_LABEL[b.quality]}</span>
					</div>
					<div class="row sub mono"><span>CN {p?.cn}</span><span>{(zone.type === 'packaging' ? b.packaged_qty : b.produced_qty).toFixed(2)} / {b.target_qty} t</span></div>
					<div class="bar"><span style="width:{Math.round(prog * 100)}%"></span></div>
					<div class="row acts">
						<button class="btn btn-sm btn-ghost" onclick={() => ws.logBatch(b.code, zone.type === 'packaging' ? 0 : 0.5, zone.type === 'packaging' ? 0.5 : 0)}>Log +0.5 t</button>
						<a class="btn btn-sm btn-ghost" href="/trace/qr?batch={b.code}">QR {b.qr_token.slice(-6)}</a>
					</div>
				</div>
			{:else}
				<p class="muted">No batches in this zone.</p>
			{/each}
			{#if zone.type === 'packaging'}
				<div class="dock card">
					<span class="eyebrow">Dispatch</span>
					<p>Container <span class="mono">ZV-001</span> linked to A-101, A-102, A-103 · ready to dispatch</p>
					{#if ws.manifest.capabilities.has('freightlink_core')}<p class="muted">Next truck: MH-14-HG-3320 · Dock 2 · ETA 14:30</p>{/if}
				</div>
			{/if}
		{/if}
	{/if}
	{#snippet footer()}
		<a class="btn btn-sm" href="/plants/plt_chakan/zones?zone={zone?.code}">Open zone detail <span class="arrow">→</span></a>
	{/snippet}
</Drawer>

<style>
	.stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1px;
		background: var(--border-hairline);
		border: var(--hairline);
		border-radius: var(--radius-md);
		overflow: hidden;
	}
	.stats > div {
		background: var(--surface-card);
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 13px;
		color: var(--ink-primary);
	}
	.big {
		font-size: 1.6rem;
		letter-spacing: -0.04em;
	}
	.big small {
		font-size: 0.5em;
		color: var(--ink-muted);
	}
	.roster {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin: 14px 0;
	}
	.roster span {
		width: 14px;
		height: 14px;
		border-radius: 3px;
		background: var(--surface-sunken);
		animation: r 400ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 20ms);
	}
	.roster span.on {
		background: var(--brand-emerald);
	}
	@keyframes r {
		from { transform: scale(0); }
	}
	.ctl {
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.stepper {
		display: flex;
		gap: 6px;
	}
	.stepper .input {
		width: 70px;
		height: 34px;
		text-align: center;
	}
	.sec {
		margin: 20px 0 10px;
		font-size: 13px;
		font-weight: 600;
	}
	.batch,
	.stock {
		padding: 12px 0;
		border-bottom: var(--hairline);
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 12.5px;
	}
	.row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
	}
	.sub {
		font-size: 11px;
		color: var(--ink-muted);
	}
	.strong {
		color: var(--ink-primary);
	}
	.acts {
		justify-content: flex-start;
	}
	.lot {
		font-size: 10.5px;
		color: var(--ink-faint);
	}
	.staff {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 6px;
	}
	.staff div {
		display: flex;
		flex-direction: column;
		padding: 10px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		font-size: 11px;
	}
	.staff .num {
		font-size: 1.3rem;
		color: var(--ink-primary);
	}
	.list {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 12.5px;
	}
	.list li {
		display: flex;
		gap: 8px;
		align-items: baseline;
	}
	.warn {
		margin-top: 10px;
		font-size: 12px;
		color: var(--status-terracotta);
	}
	.dock {
		margin-top: 14px;
		padding: 12px;
		font-size: 12.5px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
</style>
