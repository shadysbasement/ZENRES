<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Kpi from '$lib/components/ui/Kpi.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { money, num, timeAgo } from '$lib/format';

	let tab = $state<'all' | 'raw_material' | 'wip' | 'finished_good' | 'consumable'>('all');
	const tabs = [
		['all', 'All stock'],
		['raw_material', 'Raw materials'],
		['wip', 'Work in progress'],
		['finished_good', 'Finished goods'],
		['consumable', 'Consumables']
	] as const;
	const rows = $derived(tab === 'all' ? ws.inventory : ws.inventory.filter((i) => i.category === tab));
	const value = (c?: string) => ws.inventory.filter((i) => !c || i.category === c).reduce((s, i) => s + i.qty * i.unit_value_usd, 0);
	const low = $derived(ws.inventory.filter((i) => i.qty < i.reorder));

	let mv = $state({ item: 'RM-ALI-P1020', kind: 'Goods received', qty: 1, ref: '' });
	function logMovement(e: SubmitEvent) {
		e.preventDefault();
		const it = ws.inventory.find((i) => i.code === mv.item)!;
		const sign = mv.kind === 'Goods received' ? 1 : -1;
		it.qty = +(it.qty + sign * mv.qty).toFixed(3);
		ws.movements.unshift({ at: new Date().toISOString(), kind: mv.kind, item: mv.item, qty: sign * mv.qty, ref: mv.ref || '—', by: 'Ananya Deshpande' });
		ws.toast(`${mv.kind}: ${mv.qty} ${it.unit} ${it.code}`);
	}
</script>

<svelte:head><title>Inventory — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Business · A3" title="Inventory &" serif="material movement" lede="Raw materials, WIP and finished goods linked to production batches, with custody events and Tally stock-item mapping." />

<div class="kpis">
	<Kpi i={0} label="Total stock value" value={money(value(), ws.currency, { compact: true })} />
	<Kpi i={1} label="Raw materials" value={money(value('raw_material'), ws.currency, { compact: true })} />
	<Kpi i={2} label="Work in progress" value={money(value('wip'), ws.currency, { compact: true })} />
	<Kpi i={3} label="Finished goods" value={money(value('finished_good'), ws.currency, { compact: true })} />
	<Kpi i={4} label="Below reorder" value={String(low.length)} delta={low.map((l) => l.code).join(', ')} tone={low.length ? 'down' : 'flat'} />
</div>

<div class="g2">
	<section class="panel">
		<div class="tabs-line" role="tablist">
			{#each tabs as [id, label] (id)}<button role="tab" aria-selected={tab === id} onclick={() => (tab = id)}>{label}</button>{/each}
		</div>
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>Item</th><th>Category</th><th class="num">On hand</th><th>Reorder</th><th class="num">Value</th><th>Tally stock ID</th></tr></thead>
				<tbody>
					{#each rows as it (it.id)}
						{@const isLow = it.qty < it.reorder}
						<tr>
							<td><span class="strong">{it.name}</span><div class="mono small muted">{it.code}{it.lot ? ` · lot ${it.lot}` : ''}</div></td>
							<td><span class="pill">{it.category.replace('_', ' ')}</span></td>
							<td class="num" class:text-terracotta={isLow}>{num(it.qty, it.unit === 't' ? 1 : 0)} {it.unit}</td>
							<td style="min-width:120px">{#if it.reorder}<div class="bar"><span style="width:{Math.min(100, (it.qty / (it.reorder * 2)) * 100)}%;{isLow ? 'background:var(--status-terracotta)' : ''}"></span></div><span class="mono small muted">min {it.reorder}</span>{:else}<span class="muted">—</span>{/if}</td>
							<td class="num">{money(it.qty * it.unit_value_usd, ws.currency)}</td>
							<td class="mono small">{it.tally_id}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>

	<div class="side">
		<form class="panel" onsubmit={logMovement}>
			<div class="panel-head"><h2>Log movement</h2></div>
			<div class="form">
				<label class="field"><span class="label">Event</span><select class="select" bind:value={mv.kind}><option>Goods received</option><option>Issued to line</option><option>Packed into batch</option><option>Dispatched</option></select></label>
				<label class="field"><span class="label">Item</span><select class="select" bind:value={mv.item}>{#each ws.inventory as i (i.id)}<option value={i.code}>{i.code}</option>{/each}</select></label>
				<div class="two">
					<label class="field"><span class="label">Quantity</span><input class="input num" type="number" step="0.1" min="0" bind:value={mv.qty} /></label>
					<label class="field"><span class="label">Reference</span><input class="input mono" bind:value={mv.ref} placeholder="GRN / batch" /></label>
				</div>
				<button class="btn btn-sm btn-dark">Record custody event</button>
			</div>
		</form>
		<section class="panel">
			<div class="panel-head"><h2>Movement log</h2></div>
			<ol class="log" role="list">
				{#each ws.movements.slice(0, 8) as m, i (i)}
					<li><span class="mono q" class:neg={m.qty < 0}>{m.qty > 0 ? '+' : ''}{m.qty}</span><div><span class="strong">{m.kind}</span> · <span class="mono">{m.item}</span><div class="muted small">{m.ref} · {m.by} · {timeAgo(m.at)}</div></div></li>
				{/each}
			</ol>
		</section>
	</div>
</div>

<style>
	.kpis {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 12px;
		margin-bottom: 16px;
	}
	.g2 {
		display: grid;
		grid-template-columns: 1fr 340px;
		gap: 16px;
		align-items: start;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.strong {
		color: var(--ink-primary);
		font-weight: 500;
	}
	.small {
		font-size: 11px;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.two {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.log {
		display: flex;
		flex-direction: column;
	}
	.log li {
		display: grid;
		grid-template-columns: 56px 1fr;
		gap: 8px;
		padding: 9px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
	}
	.q {
		color: var(--brand-emerald);
	}
	.q.neg {
		color: var(--ink-secondary);
	}
	@media (max-width: 1200px) {
		.g2 {
			grid-template-columns: 1fr;
		}
		.kpis {
			grid-template-columns: repeat(2, 1fr);
		}
	}
</style>
