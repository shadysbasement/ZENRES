<script lang="ts">
	import { CORRIDORS, FREIGHT_STAGES, DIESEL_KG_CO2_PER_L, DISCLAIMERS } from '$lib/content';
	import { tweenNumber } from '$design/motion';
	import Segmented from '$design/components/Segmented.svelte';
	import { untrack } from 'svelte';

	interface Props {
		compact?: boolean;
	}
	let { compact = false }: Props = $props();

	let corridorIdx = $state(0);
	let stage = $state(0);
	const c = $derived(CORRIDORS[corridorIdx]);
	const s = $derived(FREIGHT_STAGES[stage]);

	// Schematic geometry (not a geographic map): cities placed relative to each other.
	const outD = $derived(curve(c.from, c.to, 1));
	const retD = $derived(curve(c.to, c.from, 1));

	function curve(a: { x: number; y: number }, b: { x: number; y: number }, side: number) {
		const mx = (a.x + b.x) / 2;
		const my = (a.y + b.y) / 2;
		const dx = b.x - a.x;
		const dy = b.y - a.y;
		const len = Math.hypot(dx, dy);
		const off = len * 0.22 * side;
		const cx = mx - (dy / len) * off;
		const cy = my + (dx / len) * off;
		return `M${a.x} ${a.y} Q ${cx.toFixed(1)} ${cy.toFixed(1)} ${b.x} ${b.y}`;
	}

	let outEl: SVGPathElement | undefined = $state();
	let retEl: SVGPathElement | undefined = $state();
	let truck = $state({ x: 0, y: 0 });
	let t = $state(0); // 0..2 along out (0..1) then return (1..2)
	let stopTween = () => {};

	// target position along the round trip for each stage
	const stageT = [0.15, 1, 1, 1, 1, 1.05, 1.55, 2];
	const loaded = $derived(stage <= 1 || stage >= 5);

	$effect(() => {
		const target = stageT[stage];
		void corridorIdx;
		untrack(() => {
			stopTween();
			stopTween = tweenNumber(t, target, 1100, (v) => (t = v));
		});
	});

	$effect(() => {
		const tv = t;
		void outD;
		if (!outEl || !retEl) return;
		const el = tv <= 1 ? outEl : retEl;
		const local = tv <= 1 ? tv : tv - 1;
		const L = el.getTotalLength();
		const p = el.getPointAtLength(Math.max(0, Math.min(1, local)) * L);
		truck = { x: p.x, y: p.y };
	});

	const saved = $derived(c.baselineFuelL - c.matchedFuelL);
	const matchedVisible = $derived(stage >= 4);

	let autoplay = $state(false);
	$effect(() => {
		if (!autoplay) return;
		const id = setInterval(() => {
			stage = (stage + 1) % FREIGHT_STAGES.length;
		}, 2200);
		return () => clearInterval(id);
	});

	function pickCorridor(i: number) {
		corridorIdx = i;
		stage = 0;
		t = 0;
	}

	const cities = [
		{ code: 'LDH', name: 'Ludhiana', x: 250, y: 40 },
		{ code: 'DEL', name: 'Delhi', x: 300, y: 110 },
		{ code: 'JAI', name: 'Jaipur', x: 255, y: 175 },
		{ code: 'AMD', name: 'Ahmedabad', x: 150, y: 235 },
		{ code: 'BOM', name: 'Mumbai', x: 150, y: 330 },
		{ code: 'PNQ', name: 'Pune', x: 190, y: 360 }
	];
</script>

<div class="fc" class:compact>
	<div class="fc-top">
		<Segmented label="Corridor" options={CORRIDORS.map((cor, i) => ({ value: String(i), label: cor.label }))} value={String(corridorIdx)} onchange={(v) => pickCorridor(Number(v))} />
		<button class="btn btn-sm btn-ghost" onclick={() => (autoplay = !autoplay)} aria-pressed={autoplay}>
			{autoplay ? '❚❚ Pause' : '▶ Play journey'}
		</button>
	</div>

	<div class="fc-body">
		<div class="map card">
			<svg viewBox="0 0 420 400" role="img" aria-label="Schematic corridor diagram for {c.label}">
				<defs>
					<pattern id="fc-grid" width="20" height="20" patternUnits="userSpaceOnUse">
						<path d="M20 0H0V20" fill="none" stroke="var(--border-hairline)" stroke-width="0.6" />
					</pattern>
				</defs>
				<rect width="420" height="400" fill="url(#fc-grid)" />
				{#each cities as city (city.code)}
					{@const onRoute = city.code === c.from.code || city.code === c.to.code}
					<g transform="translate({city.x} {city.y})" class="city" class:on={onRoute}>
						<circle r={onRoute ? 6 : 3} />
						<text x="12" y="4">{city.code}<tspan class="cname"> {city.name}</tspan></text>
					</g>
				{/each}
				{#key corridorIdx}
					<path bind:this={outEl} d={outD} class="r r-out" pathLength="1" />
					<path bind:this={retEl} d={retD} class="r r-ret" class:matched={matchedVisible} pathLength="1" />
				{/key}
				<g transform="translate({truck.x} {truck.y})" class="truck" class:loaded>
					<circle r="14" class="halo" />
					<rect x="-7" y="-5" width="14" height="10" rx="2" />
				</g>
			</svg>
			<div class="legend mono">
				<span><i class="k k-out"></i>Outbound loaded</span>
				<span><i class="k k-ret"></i>{matchedVisible ? 'Return matched' : 'Return (empty)'}</span>
				<span class="muted">Schematic — not to scale</span>
			</div>
		</div>

		<div class="readout">
			<ol class="stages" role="list">
				{#each FREIGHT_STAGES as st, i (st.key)}
					<li>
						<button class:active={stage === i} class:done={i < stage} onclick={() => { stage = i; autoplay = false; }}>
							<span class="mono">{String(i + 1).padStart(2, '0')}</span>
							<span>{st.name}</span>
						</button>
					</li>
				{/each}
			</ol>

			{#key stage}
				<div class="detail card">
					<span class="eyebrow">Stage {String(stage + 1).padStart(2, '0')} — {s.name}</span>
					<p>{s.detail}</p>
					<div class="logic mono"><span class="k">logic</span>{s.filter}</div>
				</div>
			{/key}

			<div class="facts">
				<div><span class="eyebrow">Distance</span><span class="num">{c.km.toLocaleString('en-US')} km</span></div>
				<div><span class="eyebrow">Vehicle</span><span>{c.vehicle}</span></div>
				<div><span class="eyebrow">Outbound</span><span>{c.outbound}</span></div>
				<div><span class="eyebrow">Return cargo</span><span class:muted={!matchedVisible}>{matchedVisible ? c.returnCargo : 'Awaiting match'}</span></div>
			</div>

			<div class="calc card">
				<div class="calc-row"><span>Baseline (empty return + separate trip)</span><span class="num">{c.baselineFuelL} L</span></div>
				<div class="calc-row"><span>Matched backhaul journey</span><span class="num">{c.matchedFuelL} L</span></div>
				<div class="calc-row total"><span>Illustrative avoided fuel</span><span class="num">{saved} L ≈ {Math.round(saved * DIESEL_KG_CO2_PER_L)} kg CO₂</span></div>
				<p class="fine">{DISCLAIMERS.freight}</p>
			</div>
		</div>
	</div>
</div>

<style>
	.fc-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		margin-bottom: 20px;
	}
	.fc-body {
		display: grid;
		grid-template-columns: 1.1fr 1fr;
		gap: var(--gap);
	}
	.map {
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.map svg {
		width: 100%;
		height: auto;
		border-radius: var(--radius-sm);
		background: var(--surface-canvas);
	}
	.city circle {
		fill: var(--border-strong);
		transition: r 600ms var(--ease-expo), fill 400ms;
	}
	.city.on circle {
		fill: var(--surface-card);
		stroke: var(--ink-primary);
		stroke-width: 1.6;
	}
	.city text {
		font-family: var(--font-mono);
		font-size: 10px;
		fill: var(--ink-faint);
	}
	.city.on text {
		fill: var(--ink-primary);
		font-weight: 500;
	}
	.cname {
		fill: var(--ink-muted);
		font-weight: 400;
	}
	.r {
		fill: none;
		stroke-width: 2;
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		animation: draw 1.4s var(--ease-expo) forwards;
	}
	.r-out {
		stroke: var(--ink-primary);
	}
	.r-ret {
		stroke: var(--accent-cobalt);
		opacity: 0.45;
		stroke-dasharray: 0.012 0.012;
		stroke-dashoffset: 0;
		animation: none;
		transition: opacity 500ms;
	}
	.r-ret.matched {
		stroke: var(--brand-emerald);
		opacity: 1;
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		animation: draw 1.4s var(--ease-expo) forwards;
	}
	@keyframes draw {
		to {
			stroke-dashoffset: 0;
		}
	}
	.truck rect {
		fill: var(--surface-card);
		stroke: var(--ink-primary);
		stroke-width: 1.5;
		transition: fill 400ms;
	}
	.truck.loaded rect {
		fill: var(--brand-emerald);
		stroke: var(--brand-emerald-hover);
	}
	.halo {
		fill: var(--brand-emerald);
		opacity: 0.12;
		animation: halo 2s ease-in-out infinite;
	}
	@keyframes halo {
		50% {
			opacity: 0.02;
			transform: scale(1.4);
		}
	}
	.legend {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
		font-size: 10.5px;
		color: var(--ink-secondary);
	}
	.legend .muted {
		margin-left: auto;
	}
	.k {
		display: inline-block;
		width: 14px;
		height: 2px;
		vertical-align: middle;
		margin-right: 6px;
	}
	.k-out {
		background: var(--ink-primary);
	}
	.k-ret {
		background: var(--brand-emerald);
	}
	.readout {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.stages {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 6px;
		list-style: none;
		padding: 0;
	}
	.stages button {
		width: 100%;
		height: 100%;
		min-height: 58px;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		justify-content: space-between;
		gap: 6px;
		padding: 9px 10px;
		text-align: left;
		font-size: 12px;
		line-height: 1.25;
		color: var(--ink-muted);
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		transition: all var(--dur-2) var(--ease-out);
	}
	.stages button .mono {
		font-size: 10px;
		color: var(--ink-faint);
	}
	.stages button:hover {
		border-color: var(--border-strong);
		color: var(--ink-primary);
	}
	.stages button.done {
		color: var(--ink-secondary);
		background: var(--surface-subtle);
	}
	.stages button.done .mono {
		color: var(--brand-emerald);
	}
	.stages button.active {
		background: var(--obsidian);
		border-color: var(--obsidian);
		color: var(--ink-inverse);
	}
	.stages button.active .mono {
		color: var(--vault-phosphor);
	}
	.detail {
		padding: 18px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		animation: fade 500ms var(--ease-expo);
	}
	@keyframes fade {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}
	.detail p {
		font-size: var(--fs-sm);
		color: var(--ink-primary);
		line-height: 1.55;
	}
	.logic {
		font-size: 11.5px;
		color: var(--accent-cobalt);
		background: var(--accent-cobalt-tint);
		padding: 8px 10px;
		border-radius: var(--radius-sm);
	}
	.logic .k {
		width: auto;
		height: auto;
		margin-right: 8px;
		color: var(--ink-muted);
		text-transform: uppercase;
		font-size: 10px;
	}
	.facts {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px 20px;
		padding: 4px 2px;
	}
	.facts > div {
		display: flex;
		flex-direction: column;
		gap: 3px;
		font-size: var(--fs-xs);
		color: var(--ink-primary);
	}
	.calc {
		padding: 14px 16px;
	}
	.calc-row {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		padding: 7px 0;
		font-size: var(--fs-xs);
		border-bottom: 1px dashed var(--border-hairline);
	}
	.calc-row .num {
		color: var(--ink-primary);
	}
	.calc-row.total {
		border-bottom: 0;
		font-weight: 500;
		color: var(--ink-primary);
	}
	.calc-row.total .num {
		color: var(--brand-emerald);
	}
	.fine {
		margin-top: 6px;
		font-size: 11px;
		color: var(--ink-muted);
		line-height: 1.5;
	}
	@media (max-width: 1024px) {
		.fc-body {
			grid-template-columns: 1fr;
		}
	}
	@media (max-width: 560px) {
		.stages {
			grid-template-columns: repeat(2, 1fr);
		}
		.fc-top {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
