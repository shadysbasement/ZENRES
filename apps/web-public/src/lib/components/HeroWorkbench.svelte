<script lang="ts">
	import { CBAM_MODEL, cbamExposure, CORRIDORS, DISCLAIMERS, DIESEL_KG_CO2_PER_L } from '$lib/content';
	import { tweenNumber } from '$design/motion';
	import { onMount, untrack } from 'svelte';

	let tab = $state<'cbam' | 'freight'>('cbam');
	let tonnes = $state(420);
	const ex = $derived(cbamExposure(tonnes));

	// Smoothly animated display values
	let shownDefault = $state(0);
	let shownActual = $state(0);
	let stopA = () => {};
	let stopB = () => {};
	$effect(() => {
		const d = ex.defaultEur;
		const a = ex.actualEur;
		stopA();
		stopB();
		untrack(() => {
			stopA = tweenNumber(shownDefault, d, 650, (v) => (shownDefault = v));
			stopB = tweenNumber(shownActual, a, 650, (v) => (shownActual = v));
		});
	});

	const maxEur = cbamExposure(CBAM_MODEL.maxTonnes).defaultEur;
	const eur = (v: number) => '€' + Math.round(v).toLocaleString('de-DE');
	const pct = $derived(((tonnes - CBAM_MODEL.minTonnes) / (CBAM_MODEL.maxTonnes - CBAM_MODEL.minTonnes)) * 100);

	// Freight mini preview — Delhi ↔ Mumbai
	const c = CORRIDORS[0];
	let matched = $state(false);
	let loopTimer: ReturnType<typeof setInterval>;
	onMount(() => {
		loopTimer = setInterval(() => {
			if (tab === 'freight') matched = !matched;
		}, 3600);
		return () => clearInterval(loopTimer);
	});
	$effect(() => {
		if (tab === 'freight') {
			matched = false;
			const t = setTimeout(() => (matched = true), 1400);
			return () => clearTimeout(t);
		}
	});
	const savedL = c.baselineFuelL - c.matchedFuelL;

	const outPath = 'M300 70 C 250 150, 210 230, 120 290';
	const retPath = 'M120 290 C 190 260, 270 200, 300 70';
</script>

<div class="wb card" role="region" aria-label="Interactive mini-workbench">
	<div class="wb-head">
		<div class="tabs" role="tablist">
			<button role="tab" aria-selected={tab === 'cbam'} onclick={() => (tab = 'cbam')}>
				<span class="mono">01</span> CBAM Exposure Delta
			</button>
			<button role="tab" aria-selected={tab === 'freight'} onclick={() => (tab = 'freight')}>
				<span class="mono">02</span> FreightLink Backhaul
			</button>
			<span class="tab-ink" style="transform:translateX({tab === 'cbam' ? 0 : 100}%)"></span>
		</div>
		<span class="pill pill-outline">Illustrative estimate</span>
	</div>

	{#if tab === 'cbam'}
		<div class="pane" role="tabpanel">
			<div class="row-between">
				<div>
					<span class="eyebrow">Annual EU export</span>
					<div class="big num">{tonnes.toLocaleString('en-US')}<span class="unit">t / yr</span></div>
				</div>
				<div class="product mono">{CBAM_MODEL.product}</div>
			</div>

			<input
				class="range"
				type="range"
				min={CBAM_MODEL.minTonnes}
				max={CBAM_MODEL.maxTonnes}
				step="10"
				bind:value={tonnes}
				style="--p:{pct}%"
				aria-label="Annual EU export tonnes"
			/>
			<div class="ticks mono"><span>50t</span><span>500t</span><span>1,000t</span><span>1,500t</span></div>

			<div class="bars">
				<div class="bar-row">
					<div class="bar-label">
						<span>Default-value exposure</span>
						<span class="num">{eur(shownDefault)}</span>
					</div>
					<div class="track"><span class="fill fill-default" style="width:{(ex.defaultEur / maxEur) * 100}%"></span></div>
					<span class="sub mono">{ex.defaultT.toLocaleString('en-US', { maximumFractionDigits: 0 })} tCO₂e × €{CBAM_MODEL.certificatePrice} @ {CBAM_MODEL.defaultIntensity} t/t default</span>
				</div>
				<div class="bar-row">
					<div class="bar-label">
						<span>Measured actual emissions</span>
						<span class="num">{eur(shownActual)}</span>
					</div>
					<div class="track"><span class="fill fill-actual" style="width:{(ex.actualEur / maxEur) * 100}%"></span></div>
					<span class="sub mono">{ex.actualT.toLocaleString('en-US', { maximumFractionDigits: 0 })} tCO₂e × €{CBAM_MODEL.certificatePrice} @ {CBAM_MODEL.actualIntensity} t/t measured</span>
				</div>
			</div>

			<div class="delta">
				<div>
					<span class="eyebrow">Illustrative exposure delta</span>
					<div class="delta-num num">{eur(shownDefault - shownActual)}</div>
				</div>
				<span class="pill pill-emerald"><span class="dot"></span>−{(((CBAM_MODEL.defaultIntensity - CBAM_MODEL.actualIntensity) / CBAM_MODEL.defaultIntensity) * 100).toFixed(0)}% vs default</span>
			</div>
			<p class="fine">{DISCLAIMERS.cbamEstimate}</p>
		</div>
	{:else}
		<div class="pane" role="tabpanel">
			<div class="row-between">
				<div>
					<span class="eyebrow">Corridor</span>
					<div class="big">{c.label}</div>
				</div>
				<span class="pill" class:pill-emerald={matched} class:pill-terracotta={!matched}>
					<span class="dot dot-live"></span>{matched ? 'Return load matched' : 'Searching return load'}
				</span>
			</div>
			<div class="map">
				<svg viewBox="0 0 420 340" aria-hidden="true">
					<defs>
						<pattern id="wb-dots" width="14" height="14" patternUnits="userSpaceOnUse">
							<circle cx="1" cy="1" r="0.9" fill="var(--border-hairline)" />
						</pattern>
					</defs>
					<rect width="420" height="340" fill="url(#wb-dots)" />
					<path d={outPath} class="route out" />
					<path d={retPath} class="route ret" class:on={matched} pathLength="1" />
					<circle r="5" class="truck">
						<animateMotion dur="3.6s" repeatCount="indefinite" path={outPath + ' ' + retPath.replace('M120 290', 'L120 290')} />
					</circle>
					<g class="node" transform="translate(300 70)">
						<circle r="7" /><text x="14" y="4">DEL · Delhi</text>
					</g>
					<g class="node" transform="translate(120 290)">
						<circle r="7" /><text x="-14" y="22" text-anchor="middle">BOM · Mumbai</text>
					</g>
				</svg>
				<div class="legs">
					<div class="leg">
						<span class="swatch out"></span>
						<div><span class="eyebrow">Outbound</span><p>{c.outbound}</p></div>
					</div>
					<div class="leg" class:dim={!matched}>
						<span class="swatch ret"></span>
						<div><span class="eyebrow">Matched return</span><p>{matched ? c.returnCargo : 'Empty return — 1,415 km'}</p></div>
					</div>
				</div>
			</div>
			<div class="delta">
				<div>
					<span class="eyebrow">Illustrative avoided fuel</span>
					<div class="delta-num num">{matched ? savedL : 0} L <span class="unit">≈ {matched ? Math.round(savedL * DIESEL_KG_CO2_PER_L) : 0} kg CO₂</span></div>
				</div>
				<span class="mono small">{c.baselineFuelL}L baseline → {c.matchedFuelL}L matched</span>
			</div>
			<p class="fine">{DISCLAIMERS.freight}</p>
		</div>
	{/if}
</div>

<style>
	.wb {
		padding: 0;
		box-shadow: var(--shadow-float);
		overflow: hidden;
	}
	.wb-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 10px 12px 10px 10px;
		border-bottom: var(--hairline);
		background: var(--surface-canvas);
	}
	.tabs {
		position: relative;
		display: grid;
		grid-template-columns: 1fr 1fr;
	}
	.tabs button {
		position: relative;
		z-index: 1;
		height: 34px;
		padding: 0 14px;
		font-size: var(--fs-xs);
		font-weight: 500;
		color: var(--ink-muted);
		white-space: nowrap;
		transition: color var(--dur-2) var(--ease-out);
	}
	.tabs button .mono {
		color: var(--ink-faint);
		margin-right: 4px;
	}
	.tabs button[aria-selected='true'] {
		color: var(--ink-primary);
	}
	.tabs button[aria-selected='true'] .mono {
		color: var(--brand-emerald);
	}
	.tab-ink {
		position: absolute;
		left: 0;
		top: 0;
		width: 50%;
		height: 100%;
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		transition: transform var(--dur-3) var(--ease-expo);
	}
	.pane {
		padding: 22px 22px 18px;
		animation: pane-in 600ms var(--ease-expo);
	}
	@keyframes pane-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
	}
	.row-between {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 16px;
	}
	.big {
		margin-top: 6px;
		font-size: 2.25rem;
		font-weight: 500;
		letter-spacing: -0.03em;
		color: var(--ink-primary);
		line-height: 1;
	}
	.unit {
		font-size: 0.42em;
		color: var(--ink-muted);
		margin-left: 6px;
		letter-spacing: 0;
		font-family: var(--font-mono);
	}
	.product {
		font-size: 11px;
		color: var(--ink-muted);
		text-align: right;
		max-width: 16ch;
	}
	.range {
		-webkit-appearance: none;
		appearance: none;
		width: 100%;
		height: 28px;
		margin-top: 18px;
		background: transparent;
	}
	.range::-webkit-slider-runnable-track {
		height: 2px;
		background: linear-gradient(90deg, var(--brand-emerald) var(--p), var(--border-hairline) var(--p));
	}
	.range::-moz-range-track {
		height: 2px;
		background: linear-gradient(90deg, var(--brand-emerald) var(--p), var(--border-hairline) var(--p));
	}
	.range::-webkit-slider-thumb {
		-webkit-appearance: none;
		width: 18px;
		height: 18px;
		margin-top: -8px;
		border-radius: 50%;
		background: var(--surface-card);
		border: 1.5px solid var(--brand-emerald);
		box-shadow: 0 0 0 5px var(--brand-emerald-tint);
		transition: transform var(--dur-2) var(--ease-out);
	}
	.range::-moz-range-thumb {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: var(--surface-card);
		border: 1.5px solid var(--brand-emerald);
		box-shadow: 0 0 0 5px var(--brand-emerald-tint);
	}
	.range:active::-webkit-slider-thumb {
		transform: scale(1.15);
	}
	.ticks {
		display: flex;
		justify-content: space-between;
		font-size: 10px;
		color: var(--ink-faint);
	}
	.bars {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.bar-label {
		display: flex;
		justify-content: space-between;
		font-size: var(--fs-sm);
		color: var(--ink-secondary);
	}
	.bar-label .num {
		color: var(--ink-primary);
		font-weight: 500;
	}
	.track {
		margin-top: 8px;
		height: 10px;
		background: var(--surface-subtle);
		border-radius: 2px;
		overflow: hidden;
	}
	.fill {
		display: block;
		height: 100%;
		border-radius: 2px;
		transition: width 900ms var(--ease-expo);
	}
	.fill-default {
		background: repeating-linear-gradient(135deg, var(--ink-secondary) 0 2px, var(--ink-muted) 2px 5px);
	}
	.fill-actual {
		background: var(--brand-emerald);
	}
	.sub {
		display: block;
		margin-top: 6px;
		font-size: 10.5px;
		color: var(--ink-faint);
	}
	.delta {
		margin-top: 20px;
		padding-top: 16px;
		border-top: 1px dashed var(--border-hairline);
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
	}
	.delta-num {
		margin-top: 4px;
		font-size: 1.75rem;
		font-weight: 500;
		color: var(--brand-emerald);
		letter-spacing: -0.02em;
	}
	.fine {
		margin-top: 14px;
		font-size: 11px;
		line-height: 1.5;
		color: var(--ink-muted);
	}
	.small {
		font-size: 11px;
		color: var(--ink-muted);
	}
	.map {
		margin-top: 14px;
		display: grid;
		grid-template-columns: 1.3fr 1fr;
		gap: 16px;
		align-items: center;
	}
	.map svg {
		width: 100%;
		height: auto;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		background: var(--surface-canvas);
	}
	.route {
		fill: none;
		stroke-width: 2;
	}
	.route.out {
		stroke: var(--ink-primary);
	}
	.route.ret {
		stroke: var(--accent-cobalt);
		stroke-dasharray: 0.02 0.02;
		opacity: 0.55;
		transition: stroke 600ms, opacity 600ms;
	}
	.route.ret.on {
		stroke: var(--brand-emerald);
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		opacity: 1;
		animation: draw 1.4s var(--ease-expo) forwards;
	}
	@keyframes draw {
		to {
			stroke-dashoffset: 0;
		}
	}
	.truck {
		fill: var(--brand-emerald);
		stroke: #fff;
		stroke-width: 2;
	}
	.node circle {
		fill: var(--surface-card);
		stroke: var(--ink-primary);
		stroke-width: 1.5;
	}
	.node text {
		font-family: var(--font-mono);
		font-size: 11px;
		fill: var(--ink-secondary);
	}
	.legs {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.leg {
		display: flex;
		gap: 10px;
		transition: opacity 500ms;
	}
	.leg.dim {
		opacity: 0.55;
	}
	.leg p {
		font-size: var(--fs-xs);
		color: var(--ink-primary);
		margin-top: 2px;
	}
	.swatch {
		flex: none;
		width: 14px;
		height: 2px;
		margin-top: 7px;
	}
	.swatch.out {
		background: var(--ink-primary);
	}
	.swatch.ret {
		background: var(--brand-emerald);
	}
	@media (max-width: 560px) {
		.wb-head {
			flex-direction: column;
			align-items: stretch;
		}
		.wb-head .pill {
			align-self: flex-start;
		}
		.map {
			grid-template-columns: 1fr;
		}
		.tabs button {
			padding: 0 8px;
		}
	}
</style>
