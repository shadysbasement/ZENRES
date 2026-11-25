<script lang="ts">
	import { onMount } from 'svelte';
	import { PILLARS } from '$lib/content';
	import { reveal } from '$design/motion';

	let active = $state(0);
	let progress = $state(0);
	let paused = $state(false);
	let inView = $state(false);
	let root: HTMLElement;
	const DURATION = 8000;

	onMount(() => {
		const io = new IntersectionObserver(([e]) => (inView = e.isIntersecting), { threshold: 0.3 });
		io.observe(root);
		let last = performance.now();
		let raf = 0;
		const tick = (now: number) => {
			const dt = now - last;
			last = now;
			if (inView && !paused) {
				progress += dt / DURATION;
				if (progress >= 1) {
					progress = 0;
					active = (active + 1) % PILLARS.length;
				}
			}
			raf = requestAnimationFrame(tick);
		};
		raf = requestAnimationFrame(tick);
		return () => {
			cancelAnimationFrame(raf);
			io.disconnect();
		};
	});

	function select(i: number) {
		active = i;
		progress = 0;
	}

	const p = $derived(PILLARS[active]);

	// deterministic pseudo-QR pattern
	const qr = Array.from({ length: 81 }, (_, i) => ((i * 7919 + 13) % 11) % 3 !== 0);
</script>

<div class="pillars" bind:this={root} onpointerenter={() => (paused = true)} onpointerleave={() => (paused = false)} role="presentation">
	<div class="list" role="tablist" aria-label="Four connected pillars">
		{#each PILLARS as pillar, i (pillar.index)}
			<button
				role="tab"
				class="item"
				aria-selected={active === i}
				onclick={() => select(i)}
				use:reveal={{ delay: i * 80 }}
			>
				<span class="idx mono">{pillar.index} /</span>
				<span class="name">{pillar.name}</span>
				<span class="arrow" aria-hidden="true">→</span>
				<span class="prog"><span style="transform:scaleX({active === i ? progress : 0})"></span></span>
			</button>
		{/each}
	</div>

	<div class="stage card" role="tabpanel" use:reveal={{ delay: 120 }}>
		{#key active}
			<div class="stage-inner">
				<div class="copy">
					<span class="eyebrow">{p.index} — {p.name}</span>
					<h3 class="serif">{p.headline}</h3>
					<p>{p.body}</p>
					<ul role="list" class="caps">
						{#each p.capabilities as cap, i (cap)}
							<li style="--i:{i}"><span class="tick" aria-hidden="true"></span>{cap}</li>
						{/each}
					</ul>
				</div>
				<div class="visual" aria-hidden="true">
					{#if active === 0}
						<div class="v-ledger">
							<div class="v-box mono"><span class="eyebrow">Source</span>Tally.ERP 9</div>
							<div class="v-flow"><span></span><span></span><span></span></div>
							<div class="v-rows">
								{#each [['SAL/2611/0418', 'Sales', '$18,420'], ['PUR/2611/0233', 'Purchase', '$9,870'], ['STK/2611/0091', 'Stock', '42.6 t'], ['BOM/AL-6063', 'BOM', '4 inputs']] as row, i (row[0])}
									<div class="v-row mono" style="--i:{i}"><span>{row[0]}</span><span>{row[1]}</span><span>{row[2]}</span></div>
								{/each}
							</div>
						</div>
					{:else if active === 1}
						<div class="v-carbon">
							{#each [{ l: 'Scope 1 · Fuel', v: 62, q: 'Measured' }, { l: 'Scope 2 · Grid', v: 84, q: 'Measured' }, { l: 'Precursors', v: 46, q: 'Supplier-reported' }, { l: 'Process', v: 28, q: 'Estimated' }] as b, i (b.l)}
								<div class="v-bar" style="--i:{i}">
									<div class="v-bar-top"><span>{b.l}</span><span class="pill" class:pill-emerald={b.q === 'Measured'} class:pill-cobalt={b.q === 'Supplier-reported'} class:pill-terracotta={b.q === 'Estimated'}>{b.q}</span></div>
									<div class="v-track"><span style="--w:{b.v}%"></span></div>
								</div>
							{/each}
							<div class="v-xml mono">&lt;CBAMReport rule="IR (EU) 2025/2621"&gt; … &lt;/CBAMReport&gt;</div>
						</div>
					{:else if active === 2}
						<div class="v-trace">
							<div class="v-qr">
								{#each qr as on, i (i)}<span class:on style="--i:{i}"></span>{/each}
							</div>
							<div class="v-links">
								<div class="v-chip mono strong">ZV-001</div>
								{#each ['A-101', 'A-102', 'A-103'] as b, i (b)}
									<div class="v-chip mono" style="--i:{i}"><span class="ln"></span>{b}</div>
								{/each}
							</div>
						</div>
					{:else}
						<div class="v-net">
							<svg viewBox="0 0 300 220">
								{#each [[60, 50], [40, 120], [80, 180]] as [x, y], i (i)}
									{#each [[230, 40], [250, 110], [220, 185]] as [bx, by], j (j)}
										<line x1={x} y1={y} x2={bx} y2={by} style="--i:{i * 3 + j}" />
									{/each}
								{/each}
								{#each [[60, 50], [40, 120], [80, 180]] as [x, y], i (i)}
									<g transform="translate({x} {y})"><rect x="-9" y="-9" width="18" height="18" rx="3" class="msme" /></g>
								{/each}
								{#each [[230, 40, 'EU'], [250, 110, 'UK'], [220, 185, 'US']] as [x, y, l], i (i)}
									<g transform="translate({x} {y})"><circle r="11" class="buyer" /><text y="4" text-anchor="middle">{l}</text></g>
								{/each}
							</svg>
							<div class="v-badges">
								<span class="pill pill-outline">Self-Declared</span>
								<span class="pill pill-cobalt">Document-Reviewed</span>
								<span class="pill pill-emerald">Independently Verified</span>
							</div>
						</div>
					{/if}
					<div class="metric">
						<span class="eyebrow">{p.metric.label}</span>
						<span class="num">{p.metric.value}</span>
					</div>
				</div>
			</div>
		{/key}
	</div>
</div>

<style>
	.pillars {
		display: grid;
		grid-template-columns: 5fr 7fr;
		gap: var(--gap);
		align-items: start;
	}
	.list {
		display: flex;
		flex-direction: column;
		border-top: var(--hairline);
	}
	.item {
		position: relative;
		display: grid;
		grid-template-columns: 54px 1fr auto;
		align-items: baseline;
		gap: 8px;
		padding: 26px 4px 26px 0;
		text-align: left;
		border-bottom: var(--hairline);
	}
	.idx {
		font-size: 12px;
		color: var(--ink-faint);
		transition: color var(--dur-2);
	}
	.name {
		font-size: clamp(1.35rem, 2.3vw, 2.1rem);
		font-weight: 500;
		letter-spacing: -0.025em;
		color: var(--ink-faint);
		transition: color 500ms var(--ease-out), transform 700ms var(--ease-expo);
	}
	.arrow {
		color: var(--brand-emerald);
		opacity: 0;
		transform: translateX(-10px);
		transition: opacity 400ms, transform 600ms var(--ease-expo);
	}
	.item:hover .name {
		color: var(--ink-secondary);
	}
	.item[aria-selected='true'] .name {
		color: var(--ink-primary);
		transform: translateX(6px);
	}
	.item[aria-selected='true'] .idx {
		color: var(--brand-emerald);
	}
	.item[aria-selected='true'] .arrow {
		opacity: 1;
		transform: none;
	}
	.prog {
		position: absolute;
		left: 0;
		right: 0;
		bottom: -1px;
		height: 1px;
	}
	.prog span {
		display: block;
		height: 100%;
		background: var(--brand-emerald);
		transform-origin: left;
	}
	.stage {
		position: sticky;
		top: 96px;
		min-height: 520px;
		overflow: hidden;
	}
	.stage-inner {
		display: grid;
		grid-template-columns: 1fr 1fr;
		height: 100%;
		min-height: 520px;
	}
	.copy {
		padding: 36px 32px;
		display: flex;
		flex-direction: column;
		gap: 18px;
		animation: fade-up 800ms var(--ease-expo);
	}
	.copy h3 {
		font-size: clamp(1.75rem, 2.6vw, 2.4rem);
		line-height: 1.08;
		color: var(--ink-primary);
	}
	.copy p {
		font-size: var(--fs-sm);
		line-height: 1.65;
	}
	.caps {
		display: grid;
		gap: 9px;
		margin-top: auto;
	}
	.caps li {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: var(--fs-sm);
		color: var(--ink-primary);
		animation: fade-up 700ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 50ms + 150ms);
	}
	.tick {
		width: 14px;
		height: 14px;
		flex: none;
		border-radius: 3px;
		background: var(--brand-emerald-tint)
			url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='8' height='6' fill='none'%3E%3Cpath stroke='%2314533C' stroke-width='1.5' d='m1 3 2 2 4-4'/%3E%3C/svg%3E")
			center no-repeat;
	}
	.visual {
		position: relative;
		background: var(--surface-canvas);
		border-left: var(--hairline);
		padding: 32px 28px 92px;
		display: flex;
		flex-direction: column;
		justify-content: center;
		background-image: radial-gradient(var(--border-hairline) 1px, transparent 1px);
		background-size: 16px 16px;
	}
	.metric {
		position: absolute;
		left: 28px;
		right: 28px;
		bottom: 24px;
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		padding-top: 14px;
		border-top: var(--hairline);
	}
	.metric .num {
		font-size: 1.5rem;
		color: var(--ink-primary);
	}
	@keyframes fade-up {
		from {
			opacity: 0;
			transform: translateY(14px);
		}
	}
	/* visuals */
	.v-box {
		background: var(--obsidian);
		color: var(--ink-inverse);
		border-radius: var(--radius-sm);
		padding: 12px 14px;
		font-size: 13px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		animation: fade-up 700ms var(--ease-expo);
	}
	.v-box .eyebrow {
		color: var(--vault-phosphor);
	}
	.v-flow {
		display: flex;
		justify-content: center;
		gap: 6px;
		padding: 12px 0;
	}
	.v-flow span {
		width: 4px;
		height: 4px;
		border-radius: 50%;
		background: var(--brand-emerald);
		animation: blink 1.2s infinite;
	}
	.v-flow span:nth-child(2) {
		animation-delay: 0.2s;
	}
	.v-flow span:nth-child(3) {
		animation-delay: 0.4s;
	}
	@keyframes blink {
		50% {
			opacity: 0.2;
		}
	}
	.v-rows {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.v-row {
		display: grid;
		grid-template-columns: 1.4fr 1fr auto;
		gap: 8px;
		padding: 9px 12px;
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		font-size: 11.5px;
		color: var(--ink-secondary);
		animation: fade-up 700ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 120ms + 300ms);
	}
	.v-row span:last-child {
		color: var(--ink-primary);
	}
	.v-carbon {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.v-bar-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 12px;
		color: var(--ink-primary);
	}
	.v-track {
		margin-top: 7px;
		height: 8px;
		background: var(--surface-sunken);
		border-radius: 2px;
		overflow: hidden;
	}
	.v-track span {
		display: block;
		height: 100%;
		width: var(--w);
		background: var(--brand-emerald);
		transform-origin: left;
		animation: grow 1.2s var(--ease-expo) both;
		animation-delay: calc(var(--i) * 100ms + 200ms);
	}
	.v-bar:nth-child(3) .v-track span {
		background: var(--accent-cobalt);
	}
	.v-bar:nth-child(4) .v-track span {
		background: var(--status-terracotta);
	}
	.v-bar {
		--i: 0;
	}
	@keyframes grow {
		from {
			transform: scaleX(0);
		}
	}
	.v-xml {
		font-size: 11px;
		color: var(--ink-muted);
		padding: 10px 12px;
		border: 1px dashed var(--border-strong);
		border-radius: var(--radius-sm);
	}
	.v-trace {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 20px;
		align-items: center;
	}
	.v-qr {
		display: grid;
		grid-template-columns: repeat(9, 1fr);
		gap: 3px;
		padding: 12px;
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		aspect-ratio: 1;
	}
	.v-qr span {
		border-radius: 1px;
		animation: fade-up 400ms var(--ease-out) both;
		animation-delay: calc(var(--i) * 6ms);
	}
	.v-qr span.on {
		background: var(--ink-primary);
	}
	.v-links {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.v-chip {
		position: relative;
		padding: 9px 12px;
		border: var(--hairline);
		background: var(--surface-card);
		border-radius: var(--radius-sm);
		font-size: 12px;
		animation: fade-up 600ms var(--ease-expo) both;
		animation-delay: calc(var(--i, 0) * 120ms + 200ms);
	}
	.v-chip.strong {
		background: var(--obsidian);
		color: var(--ink-inverse);
		border-color: var(--obsidian);
	}
	.v-chip:not(.strong) {
		margin-left: 18px;
	}
	.ln {
		position: absolute;
		left: -12px;
		top: 50%;
		width: 12px;
		border-top: 1px solid var(--brand-emerald);
	}
	.v-net svg {
		width: 100%;
	}
	.v-net line {
		stroke: var(--border-strong);
		stroke-width: 1;
		stroke-dasharray: 300;
		stroke-dashoffset: 300;
		animation: dash 1.6s var(--ease-expo) forwards;
		animation-delay: calc(var(--i) * 70ms);
	}
	@keyframes dash {
		to {
			stroke-dashoffset: 0;
		}
	}
	.msme {
		fill: var(--obsidian);
	}
	.buyer {
		fill: var(--surface-card);
		stroke: var(--accent-cobalt);
	}
	.v-net text {
		font-family: var(--font-mono);
		font-size: 8px;
		fill: var(--accent-cobalt);
	}
	.v-badges {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-top: 14px;
	}
	@media (max-width: 1024px) {
		.pillars {
			grid-template-columns: 1fr;
		}
		.stage {
			position: relative;
			top: 0;
		}
	}
	@media (max-width: 680px) {
		.stage-inner {
			grid-template-columns: 1fr;
		}
		.visual {
			border-left: 0;
			border-top: var(--hairline);
			min-height: 340px;
		}
	}
</style>
