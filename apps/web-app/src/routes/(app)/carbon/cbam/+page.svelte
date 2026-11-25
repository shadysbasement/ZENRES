<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { compareDefaultVsActual, specificEmbedded, article9Deduction, activityEmissions, avoidedFuel, CERT_PRICE_EUR, USD_INR, USD_EUR } from '$lib/carbon';
	import { num, QUALITY_LABEL, QUALITY_PILL } from '$lib/format';

	const full = $derived(ws.manifest.capabilities.has('precursor_tracking'));
	const art9 = $derived(ws.manifest.capabilities.has('cbam_article9_ccts'));
	let productId = $state(ws.products[0].id);
	let tonnes = $state(420);
	let price = $state(CERT_PRICE_EUR);
	let ccts = $state(780); // ₹ per tCO2e effectively paid under Indian CCTS (illustrative)
	let notes = $state('');
	let rule = $state('IR (EU) 2025/2621 · effective 2026-01-01');

	const p = $derived(ws.products.find((x) => x.id === productId)!);
	const emissions = $derived(full ? p : { ...p, precursor_tco2e_t: 0 });
	const cmp = $derived(compareDefaultVsActual(tonnes, emissions, price));
	const embeddedT = $derived(tonnes * specificEmbedded(emissions));
	const cctsEur = $derived((ccts / USD_INR) * USD_EUR);
	const deduction = $derived(art9 ? article9Deduction(embeddedT, cctsEur, price) : 0);

	const s1 = $derived(ws.activity.filter((a) => a.type === 'fuel' || a.type === 'process').reduce((s, a) => s + activityEmissions(a.qty, a.ef), 0));
	const s2 = $derived(ws.activity.filter((a) => a.type === 'electricity').reduce((s, a) => s + activityEmissions(a.qty, a.ef), 0));
	const logistics = $derived(ws.freight.shipments.reduce((s, x) => s + avoidedFuel(x.baseline_l, x.matched_l).kgCO2, 0) / 1000);

	$effect(() => {
		const n = notes;
		if (n.trim().length > 3) ws.hint('cbam', { kind: 'search_query', payload: n });
		else ws.clearHint('cbam');
	});

	const chart = $derived({
		tooltip: { trigger: 'axis' },
		grid: { top: 10, left: 4, right: 70, bottom: 0, containLabel: true },
		xAxis: { type: 'value', axisLabel: { formatter: (v: number) => `€${(v / 1000).toFixed(0)}k` } },
		yAxis: { type: 'category', data: ['Default values', 'Measured actual', ...(art9 ? ['After Art. 9'] : [])], inverse: true },
		series: [
			{
				type: 'bar',
				barWidth: 18,
				data: [
					{ value: Math.round(cmp.defaultEur), itemStyle: { color: '#9AA7A1' } },
					{ value: Math.round(cmp.actualEur), itemStyle: { color: '#14533C' } },
					...(art9 ? [{ value: Math.round(cmp.actualEur - deduction), itemStyle: { color: '#38B884' } }] : [])
				],
				itemStyle: { borderRadius: 2 },
				label: { show: true, position: 'right', fontFamily: 'IBM Plex Mono', fontSize: 11, formatter: (x: { value: number }) => `€${num(x.value)}` }
			}
		]
	});
</script>

<svelte:head><title>CBAM Engine — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Carbon · B2" title="Emissions &" serif="CBAM calculation" lede={full ? 'Full CN-code engine with precursor chains and Article 9 / CCTS deductions. Every figure is versioned and labelled for data quality.' : 'Scope 1 & 2 accounting from factory data and a default-value versus actual-emissions CBAM comparison.'}>
	{#snippet actions()}
		<select class="select rulesel" bind:value={rule}><option>IR (EU) 2025/2621 · effective 2026-01-01</option><option>IR (EU) 2023/1773 · transitional (archived)</option></select>
	{/snippet}
</PageHeader>

<div class="scope">
	<div class="panel"><span class="eyebrow">Scope 1 · direct</span><span class="big num">{num(s1, 1)}</span><span class="muted small">tCO₂e · fuel + process · Sep–Oct</span></div>
	<div class="panel"><span class="eyebrow">Scope 2 · indirect</span><span class="big num">{num(s2, 1)}</span><span class="muted small">tCO₂e · grid electricity · IN-WEST EF</span></div>
	<div class="panel"><span class="eyebrow">Embedded in selection</span><span class="big num">{num(embeddedT, 1)}</span><span class="muted small">tCO₂e · {tonnes} t {p.sku}</span></div>
	<div class="panel sep"><span class="eyebrow">Logistics (separate)</span><span class="big num">{num(logistics, 2)}</span><span class="muted small">tCO₂e avoided · never netted against CBAM</span></div>
</div>

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>Calculation workbench</h2><span class="pill pill-terracotta">Estimate until verified</span></div>
		<div class="inputs">
			<label class="field"><span class="label">CN-code SKU</span><select class="select" bind:value={productId}>{#each ws.products as x (x.id)}<option value={x.id}>{x.sku} · CN {x.cn} · {x.name}</option>{/each}</select></label>
			<label class="field"><span class="label">EU export quantity (t / yr)</span><input class="input num" type="number" min="1" bind:value={tonnes} /></label>
			<label class="field"><span class="label">Certificate price (€/tCO₂e)</span><input class="input num" type="number" min="1" bind:value={price} /></label>
		</div>
		<EChart option={chart} height={art9 ? 150 : 110} />
		<div class="lines">
			<div><span>Direct emissions</span><span class="num">{p.direct_tco2e_t} t/t</span></div>
			<div><span>Indirect emissions</span><span class="num">{p.indirect_tco2e_t} t/t</span></div>
			{#if full}<div><span>Precursor chain ({p.bom.filter((b) => b.precursor).length} inputs)</span><span class="num">{p.precursor_tco2e_t} t/t</span></div>{/if}
			<div class="tot"><span>Specific embedded emissions</span><span class="num">{specificEmbedded(emissions)} t/t</span></div>
			<div><span>Default value incl. markup</span><span class="num muted">{p.default_tco2e_t} t/t</span></div>
			<div class="tot em"><span>Illustrative exposure saved vs default</span><span class="num">€{num(cmp.savingsEur)} · −{cmp.reductionPct}%</span></div>
		</div>
		<label class="field notes"><span class="label">Calculation notes</span><textarea class="textarea" bind:value={notes} placeholder="Notes for your verifier — assumptions, allocation, evidence…"></textarea></label>
		<HintWhisper context="cbam" />
	</section>

	<div class="side">
		{#if art9}
			<section class="panel">
				<div class="panel-head"><h2>Article 9 / CCTS deduction</h2><span class="pill pill-emerald">Growth</span></div>
				<label class="field"><span class="label">Carbon price effectively paid (₹ / tCO₂e)</span><input class="input num" type="number" bind:value={ccts} /></label>
				<div class="lines">
					<div><span>≈ EUR / tCO₂e</span><span class="num">€{cctsEur.toFixed(2)}</span></div>
					<div class="tot em"><span>Deduction (capped)</span><span class="num">€{num(deduction)}</span></div>
				</div>
				<p class="fine">Subject to eligibility and evidence of the price actually paid; the deduction can never exceed the certificate obligation.</p>
			</section>
		{/if}
		<section class="panel">
			<div class="panel-head"><h2>Data quality</h2></div>
			<ul class="dq" role="list">
				{#each ws.products as x (x.id)}<li><span class="mono">{x.sku}</span><span class="pill {QUALITY_PILL[x.quality]}">{QUALITY_LABEL[x.quality]}</span></li>{/each}
			</ul>
			<p class="fine">Labels: Measured · Supplier-Reported · Estimated · Independently Verified.</p>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Next deadline</h2></div>
			<p class="small">Q4 embedded-emissions data due to <strong>Rhein Aluminium Import</strong> in 4 days.</p>
			<a class="btn btn-sm btn-dark" href="/carbon/exports">Prepare Registry export →</a>
		</section>
	</div>
</div>
<p class="fine foot">Zevra provides reporting support, not legal, tax or accounting advice. Rule set: {rule}.</p>

<style>
	.rulesel {
		height: 34px;
		width: auto;
		font-size: 12px;
	}
	.scope {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-bottom: 16px;
	}
	.scope .panel {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.scope .sep {
		background: var(--accent-cobalt-tint);
		border-color: #d2e0ea;
	}
	.big {
		font-size: 1.8rem;
		letter-spacing: -0.04em;
		color: var(--ink-primary);
	}
	.small {
		font-size: 11.5px;
	}
	.g2 {
		display: grid;
		grid-template-columns: 1.6fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.inputs {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr;
		gap: 10px;
		margin-bottom: 14px;
	}
	.lines {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.lines > div {
		display: flex;
		justify-content: space-between;
		padding: 9px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
		color: var(--ink-primary);
	}
	.lines .tot {
		font-weight: 600;
	}
	.lines .em {
		color: var(--brand-emerald);
	}
	.notes {
		margin-top: 16px;
	}
	.dq li {
		display: flex;
		justify-content: space-between;
		padding: 7px 0;
		border-bottom: var(--hairline);
		font-size: 12px;
	}
	.foot {
		margin-top: 16px;
	}
	@media (max-width: 1100px) {
		.g2,
		.inputs {
			grid-template-columns: 1fr;
		}
		.scope {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
