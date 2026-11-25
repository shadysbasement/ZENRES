<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { capexScenario } from '$lib/carbon';
	import { money, num } from '$lib/format';

	let lever = $state<'solar' | 'scrap' | 'fuel' | 'depreciation'>('solar');
	let solarKwp = $state(800);
	let scrapShare = $state(35);
	let fuelSwitch = $state(50);
	let depreciation = $state(40);
	let years = $state(10);

	const s = $derived.by(() => {
		const solarMWh = solarKwp * 1.45; // ~1,450 kWh/kWp/yr in Pune
		const solarCapex = solarKwp * 520;
		const solarSave = solarMWh * 1000 * 0.094;
		const solarT = solarMWh * 0.727;
		const scrapDelta = Math.max(0, scrapShare - 25) / 100;
		const scrapT = scrapDelta * 3.4 * 1450;
		const scrapCapex = scrapDelta > 0 ? 160000 + scrapDelta * 400000 : 0;
		const scrapSave = scrapDelta * 1450 * 220;
		const fuelT = (fuelSwitch / 100) * 61200 * 12 * 0.00202 * 0.35;
		const fuelCapex = fuelSwitch * 1800;
		const fuelSave = fuelSwitch * 900;
		return capexScenario({
			capexUsd: solarCapex + scrapCapex + fuelCapex,
			annualSavingsUsd: solarSave + scrapSave + fuelSave,
			annualAvoidedT: solarT + scrapT + fuelT,
			years,
			depreciationPct: depreciation,
			taxRatePct: 25.17
		});
	});
	const capex = $derived(-s.cumulative[0]);

	const option = $derived({
		tooltip: { trigger: 'axis' },
		grid: { top: 16, left: 4, right: 8, bottom: 0, containLabel: true },
		xAxis: { type: 'category', data: s.cumulative.map((_, i) => `Y${i}`), boundaryGap: false },
		yAxis: { type: 'value', axisLabel: { formatter: (v: number) => money(v, 'USD', { compact: true }) } },
		series: [
			{
				type: 'line',
				smooth: 0.3,
				symbolSize: 6,
				data: s.cumulative,
				lineStyle: { width: 2, color: '#14533C' },
				itemStyle: { color: '#14533C' },
				areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(20,83,60,0.18)' }, { offset: 1, color: 'rgba(20,83,60,0)' }] } },
				markLine: { symbol: 'none', lineStyle: { color: '#C05A3C', type: 'dashed' }, label: { formatter: 'break-even', color: '#C05A3C', fontFamily: 'IBM Plex Mono', fontSize: 10 }, data: [{ yAxis: 0 }] }
			}
		]
	});
</script>

<svelte:head><title>Capex Simulator — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Carbon · B5" title="Decarbonisation" serif="capex ROI simulator" lede="Model solar PV, scrap-ratio optimisation, fuel switching and qualifying-asset depreciation — and see payback, IRR, avoided tCO₂e and projected CBAM savings." />

<div class="g2">
	<section class="panel">
		<div class="tabs-line" role="tablist">
			{#each [['solar', 'Solar PV'], ['scrap', 'Scrap ratio'], ['fuel', 'Fuel switching'], ['depreciation', 'Depreciation']] as [id, l] (id)}
				<button role="tab" aria-selected={lever === id} onclick={() => (lever = id as typeof lever)}>{l}</button>
			{/each}
		</div>
		<div class="sliders">
			<label class:hl={lever === 'solar'}><span>Rooftop / open-access solar</span><strong class="num">{num(solarKwp)} kWp</strong><input type="range" min="0" max="3000" step="50" bind:value={solarKwp} /></label>
			<label class:hl={lever === 'scrap'}><span>Post-consumer scrap share</span><strong class="num">{scrapShare}%</strong><input type="range" min="10" max="70" bind:value={scrapShare} /></label>
			<label class:hl={lever === 'fuel'}><span>Furnace gas → electric induction</span><strong class="num">{fuelSwitch}%</strong><input type="range" min="0" max="100" step="5" bind:value={fuelSwitch} /></label>
			<label class:hl={lever === 'depreciation'}><span>Accelerated depreciation (yr 1)</span><strong class="num">{depreciation}%</strong><input type="range" min="0" max="60" step="5" bind:value={depreciation} /></label>
			<label><span>Horizon</span><strong class="num">{years} yrs</strong><input type="range" min="5" max="20" bind:value={years} /></label>
		</div>
	</section>
	<section class="panel">
		<div class="res">
			<div><span class="eyebrow">Capex</span><span class="num big">{money(capex, ws.currency, { compact: true })}</span></div>
			<div><span class="eyebrow">Payback</span><span class="num big">{s.payback ?? '—'} <small>yrs</small></span></div>
			<div><span class="eyebrow">IRR</span><span class="num big text-emerald">{s.irr ?? '—'}%</span></div>
			<div><span class="eyebrow">Avoided over horizon</span><span class="num big">{num(s.avoidedT)} <small>tCO₂e</small></span></div>
			<div><span class="eyebrow">Projected CBAM savings / yr</span><span class="num big">{money(s.cbamSavingsUsd, ws.currency, { compact: true })}</span></div>
			<div><span class="eyebrow">Total annual benefit</span><span class="num big">{money(s.yearly, ws.currency, { compact: true })}</span></div>
		</div>
		<EChart {option} height={230} />
		<div class="ref">
			<p class="small">Want a site survey? Vetted solar partners in Maharashtra can quote your rooftop.</p>
			<button class="btn btn-sm btn-dark" onclick={() => ws.toast('Solar partner referral requested · disclosed referral')}>Request solar partner referral</button>
		</div>
		<p class="fine">Illustrative model: Pune irradiance ~1,450 kWh/kWp/yr, grid EF 0.727 kg/kWh, 25.17% tax rate, €76 certificate price, 70% EU export share.</p>
	</section>
</div>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1fr 1.3fr;
		gap: 16px;
		align-items: start;
	}
	.sliders {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.sliders label {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 4px 12px;
		padding: 12px;
		border-radius: var(--radius-sm);
		font-size: 13px;
		transition: background 300ms;
	}
	.sliders label.hl {
		background: var(--brand-emerald-tint);
	}
	.sliders input {
		grid-column: 1 / -1;
		accent-color: var(--brand-emerald);
	}
	.res {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 14px;
		margin-bottom: 14px;
	}
	.res > div {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.big {
		font-size: 1.5rem;
		letter-spacing: -0.04em;
		color: var(--ink-primary);
	}
	.big small {
		font-size: 0.55em;
		color: var(--ink-muted);
	}
	.ref {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-top: var(--hairline);
		margin-top: 12px;
	}
	.small {
		font-size: 12px;
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
	}
</style>
