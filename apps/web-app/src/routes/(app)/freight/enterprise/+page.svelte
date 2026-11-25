<script lang="ts">
	import { onMount } from 'svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Kpi from '$lib/components/ui/Kpi.svelte';

	let t = $state(0);
	onMount(() => {
		const id = setInterval(() => (t += 1), 1200);
		return () => clearInterval(id);
	});
	const fleet = [
		{ reg: 'MH-04-KT-7781', route: 'Mumbai → Delhi', x0: 150, y0: 330, x1: 300, y1: 110, load: 92, status: 'Loaded' },
		{ reg: 'HR-55-AL-2290', route: 'Delhi → Mumbai', x0: 300, y0: 110, x1: 150, y1: 330, load: 64, status: 'Backhaul' },
		{ reg: 'PB-10-GH-4412', route: 'Delhi → Ludhiana', x0: 300, y0: 110, x1: 250, y1: 40, load: 78, status: 'Loaded' },
		{ reg: 'GJ-12-BX-1180', route: 'Mundra → Ahmedabad', x0: 60, y0: 230, x1: 150, y1: 235, load: 0, status: 'Empty' },
		{ reg: 'MH-14-HG-3320', route: 'Chakan → JNPT', x0: 190, y0: 360, x1: 140, y1: 345, load: 86, status: 'Loaded' }
	];
	const recs = [
		{ title: 'Pre-position GJ-12-BX-1180 at Ahmedabad for Friday Mundra exports', conf: 0.87, saving: '₹18,400 · 210 empty km' },
		{ title: 'Consolidate Vizag + Nagpur LCL into one 40 ft to JNPT', conf: 0.74, saving: '₹42,000 · 1 truck' },
		{ title: 'Shift Ludhiana pickups to Tuesday to meet Delhi backhaul window', conf: 0.66, saving: '₹9,800 / week' }
	];
	const pos = (f: (typeof fleet)[number], i: number) => {
		const p = ((t * 0.04 + i * 0.21) % 1);
		return { x: f.x0 + (f.x1 - f.x0) * p, y: f.y0 + (f.y1 - f.y0) * p };
	};
</script>

<svelte:head><title>Enterprise Fleet — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Trace & FreightLink · C5" title="Enterprise fleet &" serif="international coordination" lede="Live multi-truck fleet, bulk and AI-assisted backhaul, cargo consolidation and multi-company freight performance — for EU Importer Hub networks." />

<div class="kpis">
	<Kpi i={0} label="Trucks live" value="5" delta="consent-based GPS" />
	<Kpi i={1} label="Payload utilisation" value="81%" delta="+5 pts vs Oct" tone="up" />
	<Kpi i={2} label="Empty km" value="22%" delta="−9 pts" tone="up" />
	<Kpi i={3} label="Green logistics score" value="B+" delta="network median C" tone="up" />
</div>

<div class="g2">
	<section class="panel dark">
		<div class="panel-head"><h2>Live multi-truck fleet</h2><span class="mono small">schematic · not to scale</span></div>
		<svg viewBox="0 0 400 390" class="map">
			{#each fleet as f (f.reg)}<line x1={f.x0} y1={f.y0} x2={f.x1} y2={f.y1} class="r" />{/each}
			{#each [[250, 40, 'LDH'], [300, 110, 'DEL'], [150, 235, 'AMD'], [60, 230, 'MUN'], [150, 330, 'BOM'], [140, 345, 'JNPT'], [190, 360, 'PNQ']] as [x, y, l] (l)}
				<g transform="translate({x} {y})"><circle r="4" class="c" /><text x="8" y="3">{l}</text></g>
			{/each}
			{#each fleet as f, i (f.reg)}
				{@const p = pos(f, i)}
				<g transform="translate({p.x} {p.y})" class="truck"><circle r="9" class="halo" /><circle r="4.5" class:empty={f.load === 0} /></g>
			{/each}
		</svg>
		<ul class="fl" role="list">{#each fleet as f (f.reg)}<li><span class="mono">{f.reg}</span><span>{f.route}</span><span class="mono">{f.load}%</span></li>{/each}</ul>
	</section>
	<div class="side">
		<section class="panel">
			<div class="panel-head"><h2>AI predictive backhaul</h2><span class="pill pill-outline">model v0.4</span></div>
			{#each recs as r (r.title)}
				<div class="rec">
					<p>{r.title}</p>
					<div class="row"><span class="mono small">{r.saving}</span><span class="conf" class:low={r.conf < 0.7}>confidence {(r.conf * 100).toFixed(0)}%</span></div>
					<div class="bar"><span style="width:{r.conf * 100}%"></span></div>
				</div>
			{/each}
			<p class="fine">Recommendations are suggestions with model-confidence metrics; bookings still require explicit acceptance.</p>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Cargo consolidation planner</h2></div>
			<div class="cons">
				{#each [['Vizag Aluminium Foils', 9.2], ['Nagpur Alloy Castings', 6.8], ['Chennai Cast Aluminium', 7.2]] as [n, t] (n)}
					<div class="cl"><span>{n}</span><span class="mono">{t} t</span></div>
				{/each}
				<div class="cl tot"><span>40 ft HC to JNPT → Rotterdam</span><span class="mono">23.2 / 26 t</span></div>
			</div>
			<p class="fine">International coordination connects carriers, freight forwarders and export documents. ZENRES does not operate vessels or perform customs clearance.</p>
		</section>
	</div>
</div>

<style>
	.kpis {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-bottom: 16px;
	}
	.g2 {
		display: grid;
		grid-template-columns: 1.3fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.dark {
		background: var(--obsidian);
		border-color: var(--obsidian-line);
		color: #b9c6c0;
	}
	.dark h2 {
		color: var(--ink-inverse);
	}
	.map {
		width: 100%;
		max-height: 420px;
	}
	.r {
		stroke: var(--obsidian-line);
		stroke-width: 1.5;
		stroke-dasharray: 3 4;
	}
	.c {
		fill: var(--obsidian-3);
		stroke: #5d7268;
	}
	text {
		font-family: var(--font-mono);
		font-size: 9px;
		fill: #8fa39a;
	}
	.truck circle:last-child {
		fill: var(--vault-phosphor);
		transition: all 1.2s linear;
	}
	.truck circle.empty {
		fill: var(--status-terracotta);
	}
	.halo {
		fill: var(--vault-phosphor);
		opacity: 0.15;
	}
	.truck {
		transition: transform 1.2s linear;
	}
	.fl li {
		display: grid;
		grid-template-columns: 120px 1fr auto;
		gap: 8px;
		padding: 7px 0;
		border-top: 1px solid var(--obsidian-line);
		font-size: 12px;
	}
	.rec {
		padding: 10px 0;
		border-bottom: var(--hairline);
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13px;
		color: var(--ink-primary);
	}
	.row {
		display: flex;
		justify-content: space-between;
	}
	.small {
		font-size: 11.5px;
	}
	.conf {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--brand-emerald);
	}
	.conf.low {
		color: var(--status-terracotta);
	}
	.cl {
		display: flex;
		justify-content: space-between;
		padding: 8px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
	}
	.cl.tot {
		font-weight: 600;
		color: var(--brand-emerald);
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
		.kpis {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
