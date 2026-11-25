<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import Drawer from '$lib/components/ui/Drawer.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { money, QUALITY_LABEL, QUALITY_PILL } from '$lib/format';
	import { specificEmbedded } from '$lib/carbon';

	let open = $state(false);
	let f = $state({ sku: '', cn: '', name: '', sector: 'Aluminium' });
	const cnValid = $derived(/^\d{8}$/.test(f.cn));
	const max = $derived(ws.manifest.limits.max_cn_skus);

	function start() {
		if (max !== null && ws.products.length >= max) {
			ws.hint('products', { kind: 'limit_hit', payload: 'add_sku', count: ws.products.length });
			return;
		}
		open = true;
	}
	function save() {
		ws.products.push({ id: `prd_${Date.now()}`, sku: f.sku, cn: f.cn, name: f.name, sector: f.sector, unit_weight_kg: 1000, capacity_tpm: 0, unit_cost_usd: 0, wastage_pct: 0, cost: { material: 0, labour: 0, energy: 0, overhead: 0 }, bom: [], published_connect: false, direct_tco2e_t: 0, indirect_tco2e_t: 0, precursor_tco2e_t: 0, default_tco2e_t: 7.9, quality: 'estimated' });
		open = false;
		ws.toast(`${f.sku} mapped to CN ${f.cn}`);
		f = { sku: '', cn: '', name: '', sector: 'Aluminium' };
	}
</script>

<svelte:head><title>Products & SKUs — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Business · A2" title="Products &" serif="CN-code SKUs" lede="Export products mapped to 8-digit EU Combined Nomenclature codes, with bills of materials, costing and embedded emissions.">
	{#snippet actions()}
		<span class="mono muted usage">{ws.products.length}{max ? ` / ${max}` : ''} SKUs</span>
		<button class="btn btn-sm btn-primary" onclick={start}><Icon name="plus" size={14} />Map new SKU</button>
	{/snippet}
</PageHeader>

<HintWhisper context="products" />

<div class="cards">
	{#each ws.products as p, i (p.id)}
		{@const see = specificEmbedded(p, { includePrecursors: ws.manifest.capabilities.has('precursor_tracking') })}
		<a class="panel prod fade-in" href="/products/{p.id}" style="--i:{i}">
			<div class="top">
				<span class="mono sku">{p.sku}</span>
				<span class="pill {QUALITY_PILL[p.quality]}">{QUALITY_LABEL[p.quality]}</span>
			</div>
			<h3>{p.name}</h3>
			<span class="mono cn">CN {p.cn.slice(0, 4)} {p.cn.slice(4, 6)} {p.cn.slice(6)}</span>
			<div class="kv">
				<div><span class="eyebrow">Embedded</span><span class="num">{see.toFixed(2)} <small>tCO₂e/t</small></span></div>
				<div><span class="eyebrow">Default</span><span class="num muted">{p.default_tco2e_t} <small>t/t</small></span></div>
				<div><span class="eyebrow">Unit cost</span><span class="num">{money(p.unit_cost_usd, ws.currency)}</span></div>
				<div><span class="eyebrow">Capacity</span><span class="num">{p.capacity_tpm} <small>t/mo</small></span></div>
			</div>
			<div class="foot">
				<span class="muted">{p.bom.length} BOM inputs</span>
				{#if p.published_connect}<span class="pill pill-cobalt">On Zevra Connect</span>{/if}
			</div>
		</a>
	{/each}
</div>

<Drawer open={open} title="Map a new export SKU" eyebrow="Products" onclose={() => (open = false)}>
	<div class="form">
		<label class="field"><span class="label">SKU code</span><input class="input mono" bind:value={f.sku} placeholder="ALT-3003" /></label>
		<label class="field"><span class="label">Product name</span><input class="input" bind:value={f.name} placeholder="Aluminium tubes 3003" /></label>
		<label class="field"><span class="label">EU CN code (8 digits)</span><input class="input mono" bind:value={f.cn} maxlength="8" aria-invalid={f.cn.length > 0 && !cnValid} placeholder="76083000" /><span class="hint">CBAM goods: iron & steel (72–73), aluminium (76), cement (2523), fertilisers (28, 31), hydrogen (2804 10).</span></label>
		<label class="field"><span class="label">Sector</span><select class="select" bind:value={f.sector}><option>Aluminium</option><option>Iron & Steel</option><option>Cement</option><option>Fertilisers</option><option>Hydrogen</option></select></label>
	</div>
	{#snippet footer()}
		<button class="btn btn-sm btn-primary" disabled={!f.sku || !cnValid || !f.name} onclick={save}>Map SKU</button>
	{/snippet}
</Drawer>

<style>
	.usage {
		font-size: 11.5px;
		align-self: center;
	}
	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 14px;
	}
	.prod {
		display: flex;
		flex-direction: column;
		gap: 8px;
		transition: border-color 250ms, transform 500ms var(--ease-expo), box-shadow 300ms;
	}
	.prod:hover {
		border-color: var(--border-strong);
		transform: translateY(-3px);
		box-shadow: var(--shadow-float);
	}
	.top {
		display: flex;
		justify-content: space-between;
	}
	.sku {
		font-size: 11px;
		color: var(--brand-emerald);
	}
	h3 {
		font-size: 1.15rem;
	}
	.cn {
		font-size: 12px;
		color: var(--ink-muted);
	}
	.kv {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
		padding: 12px 0;
		border-block: var(--hairline);
		margin-top: 6px;
	}
	.kv > div {
		display: flex;
		flex-direction: column;
		gap: 3px;
		color: var(--ink-primary);
	}
	small {
		font-size: 0.75em;
		color: var(--ink-muted);
	}
	.foot {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 12px;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
</style>
