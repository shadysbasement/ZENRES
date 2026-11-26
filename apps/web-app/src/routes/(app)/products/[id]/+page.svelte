<script lang="ts">
	import { page } from '$app/state';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { money, num, QUALITY_LABEL, QUALITY_PILL } from '$lib/format';
	import { specificEmbedded, compareDefaultVsActual } from '$lib/carbon';

	const p = $derived(ws.products.find((x) => x.id === page.params.id));
	const precursor = $derived(ws.manifest.capabilities.has('precursor_tracking'));
	const batches = $derived(p ? ws.batches.filter((b) => b.product_id === p.id) : []);
	const costTotal = $derived(p ? p.cost.material + p.cost.labour + p.cost.energy + p.cost.overhead : 0);
	let qty = $state(10);
	const cmp = $derived(p ? compareDefaultVsActual(100, p) : null);

	const costOption = $derived(
		p && {
			tooltip: { trigger: 'item' },
			series: [
				{
					type: 'pie',
					radius: ['58%', '82%'],
					padAngle: 1.5,
					itemStyle: { borderRadius: 2 },
					label: { show: false },
					data: [
						{ name: 'Material', value: p.cost.material, itemStyle: { color: '#47544F' } },
						{ name: 'Energy', value: p.cost.energy, itemStyle: { color: '#8CBDE9' } },
						{ name: 'Labour', value: p.cost.labour, itemStyle: { color: '#B6A0DD' } },
						{ name: 'Overhead', value: p.cost.overhead, itemStyle: { color: '#CFDEED' } }
					]
				}
			]
		}
	);
</script>

<svelte:head><title>{p ? p.sku : 'Product'} — ZENRES</title></svelte:head>

{#if p}
	<PageHeader eyebrow="Products · CN {p.cn}" title={p.name} lede="{p.sector} · capacity {p.capacity_tpm} t/month · wastage {p.wastage_pct}%">
		{#snippet actions()}
			<span class="pill {QUALITY_PILL[p.quality]}">{QUALITY_LABEL[p.quality]}</span>
			<label class="pub"><input type="checkbox" bind:checked={p.published_connect} /> Publish to ZENRES Connect</label>
		{/snippet}
	</PageHeader>

	<div class="g3">
		<section class="panel">
			<div class="panel-head"><h2>Bill of materials</h2>{#if precursor}<span class="pill pill-emerald">Precursor tracking</span>{/if}</div>
			<div class="table-wrap">
				<table class="table">
					<thead><tr><th>Input</th><th class="num">Qty / t</th>{#if precursor}<th>Supplier</th><th class="num">tCO₂e/t</th>{/if}</tr></thead>
					<tbody>
						{#each p.bom as b (b.input)}
							<tr>
								<td class="strong">{b.input} {#if b.precursor}<span class="pill pill-outline">precursor</span>{/if}</td>
								<td class="num">{b.qty_per_t} {b.unit}</td>
								{#if precursor}<td>{b.supplier ?? '—'}</td><td class="num">{b.tco2e_per_t ?? '—'}</td>{/if}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		<section class="panel">
			<div class="panel-head"><h2>Production cost / t</h2><span class="num strong">{money(costTotal, ws.currency)}</span></div>
			<div class="pie">
				{#if costOption}<EChart option={costOption} height={170} />{/if}
				<ul role="list">
					{#each Object.entries(p.cost) as [k, v] (k)}
						<li><span class="cap">{k}</span><span class="num">{money(v, ws.currency)}</span><span class="mono muted">{((v / costTotal) * 100).toFixed(0)}%</span></li>
					{/each}
				</ul>
			</div>
			<div class="calc">
				<span class="label">Unit cost calculator</span>
				<div class="row"><input class="input num" type="number" bind:value={qty} min="1" /><span class="muted">tonnes →</span><strong class="num">{money(costTotal * qty * (1 + p.wastage_pct / 100), ws.currency)}</strong></div>
				<span class="hint">Includes {p.wastage_pct}% wastage.</span>
			</div>
		</section>

		<section class="panel">
			<div class="panel-head"><h2>Embedded emissions / t</h2></div>
			{#each [['Direct (Scope 1)', p.direct_tco2e_t], ['Indirect (Scope 2)', p.indirect_tco2e_t], ...(precursor ? [['Precursors', p.precursor_tco2e_t]] : [])] as [l, v] (l)}
				<div class="em"><span>{l}</span><span class="num">{(v as number).toFixed(2)}</span></div>
			{/each}
			<div class="em total"><span>Specific embedded</span><span class="num">{specificEmbedded(p, { includePrecursors: precursor }).toFixed(2)} tCO₂e/t</span></div>
			<div class="em"><span class="muted">Default value (incl. markup)</span><span class="num muted">{p.default_tco2e_t}</span></div>
			{#if cmp}<p class="fine">Per 100 t exported, measured data avoids an illustrative €{num(cmp.savingsEur)} versus default values.</p>{/if}
		</section>
	</div>

	<section class="panel" style="margin-top:16px">
		<div class="panel-head"><h2>Batches</h2></div>
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>Batch</th><th>Stage</th><th>Produced</th><th class="num">Target</th><th class="num">S1+S2 tCO₂e</th><th>QR token</th></tr></thead>
				<tbody>
					{#each batches as b (b.id)}
						<tr><td class="strong mono">{b.code}</td><td><span class="pill">{b.stage.replace('_', ' ')}</span></td><td>{b.produced_on}</td><td class="num">{b.target_qty} t</td><td class="num">{(b.direct_tco2e + b.indirect_tco2e).toFixed(2)}</td><td class="mono small">{b.qr_token}</td></tr>
					{:else}<tr><td colspan="6" class="empty">No batches yet.</td></tr>{/each}
				</tbody>
			</table>
		</div>
	</section>
{:else}
	<div class="empty">This SKU does not exist in your workspace. <a class="link-u" href="/products">Back to products</a></div>
{/if}

<style>
	.g3 {
		display: grid;
		grid-template-columns: 1.3fr 1fr 1fr;
		gap: 16px;
	}
	.pub {
		display: flex;
		gap: 8px;
		align-items: center;
		font-size: 12.5px;
		accent-color: var(--brand-emerald);
	}
	.strong {
		color: var(--ink-primary);
		font-weight: 500;
	}
	.small {
		font-size: 11px;
	}
	.pie {
		display: grid;
		grid-template-columns: 150px 1fr;
		align-items: center;
		gap: 10px;
	}
	.pie ul {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 12.5px;
	}
	.pie li {
		display: grid;
		grid-template-columns: 1fr auto 40px;
		gap: 8px;
	}
	.cap {
		text-transform: capitalize;
	}
	.calc {
		margin-top: 14px;
		padding-top: 14px;
		border-top: var(--hairline);
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.calc .row {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.calc .input {
		width: 90px;
		height: 34px;
	}
	.em {
		display: flex;
		justify-content: space-between;
		padding: 10px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
		color: var(--ink-primary);
	}
	.em.total {
		font-weight: 600;
		color: var(--brand-emerald);
	}
	.fine {
		margin-top: 10px;
	}
	@media (max-width: 1200px) {
		.g3 {
			grid-template-columns: 1fr;
		}
	}
</style>
