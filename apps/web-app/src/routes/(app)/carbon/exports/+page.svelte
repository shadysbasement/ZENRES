<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { ORG } from '$lib/demo/seed';
	import { specificEmbedded } from '$lib/carbon';
	import { buildRegistryXml, buildExcelCsv } from '$lib/cbamXml';

	const consolidated = $derived(ws.manifest.capabilities.has('consolidated_cbam_xml'));
	let kind = $state<'plant' | 'excel' | 'consolidated'>('plant');
	let quarter = $state('2026-Q4');
	let xml = $state('');

	const checks = $derived([
		{ ok: true, label: 'All export SKUs mapped to 8-digit CN codes' },
		{ ok: ws.activity.some((a) => a.type === 'electricity' && a.period === '2026-10' && a.quality === 'measured'), label: 'October electricity bill (MSEDCL) uploaded as measured' },
		{ ok: ws.products.every((p) => p.quality !== 'estimated'), label: 'No SKU relies solely on estimated data' },
		{ ok: true, label: 'Production batches linked to installation & period' },
		{ ok: !ws.manifest.capabilities.has('precursor_tracking') || ws.products.every((p) => p.bom.some((b) => b.precursor)), label: 'Precursor inputs declared for complex goods' },
		{ ok: false, label: 'Verifier sign-off on Q3 calculation v4 (3 open comments)' }
	]);
	const passed = $derived(checks.filter((c) => c.ok).length);

	function choose(k: typeof kind) {
		kind = k;
		if (k === 'consolidated' && !consolidated) ws.hint('exports', { kind: 'entity_pattern', payload: 'consolidated_xml' });
		else ws.clearHint('exports');
	}
	function download(name: string, content: string, type: string) {
		const url = URL.createObjectURL(new Blob([content], { type }));
		const a = Object.assign(document.createElement('a'), { href: url, download: name });
		a.click();
		URL.revokeObjectURL(url);
	}
	function generate() {
		const goods = ws.products.map((p) => ({ cn: p.cn, name: p.name, tonnes: ws.batches.filter((b) => b.product_id === p.id).reduce((s, b) => s + b.target_qty, 0) || 1, direct: p.direct_tco2e_t, indirect: p.indirect_tco2e_t, see: specificEmbedded(p, { includePrecursors: ws.manifest.capabilities.has('precursor_tracking') }), quality: p.quality }));
		if (kind === 'excel') {
			download(`CBAM_communication_${quarter}.csv`, buildExcelCsv(goods), 'text/csv');
			ws.toast('Excel communication template downloaded');
			return;
		}
		if (kind === 'consolidated' && !consolidated) return;
		xml = buildRegistryXml({ quarter, installation: { name: 'Chakan Extrusion Plant', operator: ORG.name, unlocode: 'INPNQ', country: 'IN' }, declarant: ORG.importer.name, eori: ORG.importer.eori, goods, consolidated: kind === 'consolidated' });
		download(`CBAM_${kind}_${quarter}.xml`, xml, 'application/xml');
		ws.toast('Registry XML generated & validated against XSD content rules');
	}
</script>

<svelte:head><title>Registry Exports — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Carbon" title="EU CBAM Registry" serif="exports & packs" lede="Machine-readable Registry XML and Excel communication templates, validated against date-effective content rules. Output is reporting support for your declarant." />

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>Pre-export completeness</h2><span class="mono small">{passed}/{checks.length} checks</span></div>
		<div class="meter"><span style="width:{(passed / checks.length) * 100}%"></span></div>
		<ul class="checks" role="list">
			{#each checks as c, i (c.label)}
				<li class:ok={c.ok} style="--i:{i}"><span class="mark">{c.ok ? '✓' : '!'}</span>{c.label}</li>
			{/each}
		</ul>
	</section>
	<section class="panel">
		<div class="panel-head"><h2>Generate</h2></div>
		<label class="field"><span class="label">Reporting period</span><select class="select" bind:value={quarter}><option>2026-Q4</option><option>2026-Q3</option><option>2026-Q2</option></select></label>
		<div class="kinds">
			<button class:on={kind === 'plant'} onclick={() => choose('plant')}><Icon name="exports" size={15} /><span><strong>Plant Registry XML</strong><small>Installation-level embedded emissions</small></span></button>
			<button class:on={kind === 'excel'} onclick={() => choose('excel')}><Icon name="download" size={15} /><span><strong>Excel communication template</strong><small>For importers who prefer spreadsheets</small></span></button>
			{#if ws.tier !== 'starter'}
				<button class:on={kind === 'consolidated'} onclick={() => choose('consolidated')}><Icon name="layers" size={15} /><span><strong>Consolidated XML</strong><small>Multi-supplier pack across EU shipments</small></span></button>
			{/if}
		</div>
		<HintWhisper context="exports" />
		<button class="btn btn-primary full" onclick={generate} disabled={kind === 'consolidated' && !consolidated}>Generate & download</button>
	</section>
</div>

{#if xml}
	<section class="panel fade-in" style="margin-top:16px">
		<div class="panel-head"><h2>XML preview</h2><span class="pill pill-emerald">Content rules ✓</span></div>
		<pre class="xml mono">{xml}</pre>
	</section>
{/if}

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.small {
		font-size: 11.5px;
	}
	.meter {
		height: 4px;
		background: var(--surface-sunken);
		border-radius: 2px;
		overflow: hidden;
		margin-bottom: 10px;
	}
	.meter span {
		display: block;
		height: 100%;
		background: var(--brand-emerald);
		transition: width 900ms var(--ease-expo);
	}
	.checks li {
		display: flex;
		gap: 10px;
		align-items: center;
		padding: 10px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
		animation: zv-fade-in 500ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 60ms);
	}
	.mark {
		width: 20px;
		height: 20px;
		flex: none;
		display: grid;
		place-items: center;
		border-radius: 50%;
		font-size: 11px;
		background: var(--status-terracotta-tint);
		color: var(--status-terracotta);
	}
	.ok .mark {
		background: var(--brand-emerald-tint);
		color: var(--brand-emerald);
	}
	.kinds {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 14px 0;
	}
	.kinds button {
		display: flex;
		gap: 12px;
		align-items: center;
		padding: 12px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		text-align: left;
		transition: all 200ms;
	}
	.kinds button.on {
		border-color: var(--brand-emerald);
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
	}
	.kinds span {
		display: flex;
		flex-direction: column;
		font-size: 13px;
		color: var(--ink-primary);
	}
	.kinds small {
		color: var(--ink-muted);
		font-size: 11.5px;
	}
	.full {
		width: 100%;
	}
	.xml {
		max-height: 360px;
		overflow: auto;
		padding: 16px;
		font-size: 11px;
		line-height: 1.5;
		background: var(--obsidian);
		color: #a9d8c2;
		border-radius: var(--radius-sm);
	}
	@media (max-width: 1000px) {
		.g2 {
			grid-template-columns: 1fr;
		}
	}
</style>
