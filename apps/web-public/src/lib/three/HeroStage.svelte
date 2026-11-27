<script lang="ts">
	import { onMount } from 'svelte';
	import type { Component } from 'svelte';
	import { webglAvailable } from './shared';
	import type { HotspotId } from './HeroScene.svelte';

	interface Props {
		play: boolean;
	}
	let { play }: Props = $props();

	const HOTSPOTS: Record<HotspotId, { n: string; pillar: string; title: string; body: string; stat: [string, string]; href: string }> = {
		factory: {
			n: '01',
			pillar: 'ZENRES Business · Carbon',
			title: 'Your plant, measured',
			body: 'Tally vouchers, meters and floor sheets stream into Scope 1 & 2 accounting — every figure labelled Measured, Supplier-Reported or Estimated.',
			stat: ['5.50', 'tCO₂e/t measured vs 7.9 default'],
			href: '#cbam'
		},
		container: {
			n: '02',
			pillar: 'ZENRES Trace',
			title: 'A passport for every batch',
			body: 'Container ZV-001 links batches A-101 to A-103. Buyers scan once and see only what their role permits.',
			stat: ['3', 'batches linked · hash-anchored'],
			href: '#traceability'
		},
		freight: {
			n: '03',
			pillar: 'FreightLink',
			title: 'No more empty return trips',
			body: 'Return capacity on Delhi ↔ Mumbai is matched to load demand with hard filters first — and never auto-booked.',
			stat: ['−40 L', 'illustrative fuel avoided per trip'],
			href: '#freightlink'
		},
		port: {
			n: '04',
			pillar: 'ZENRES Connect',
			title: 'Discoverable to EU buyers',
			body: 'Carbon-enriched catalogues and Registry-ready XML go out with every shipment your importer files.',
			stat: ['15–500', 'suppliers per Importer Hub'],
			href: '#platform'
		}
	};

	let host: HTMLDivElement;
	let Canvas: Component<Record<string, unknown>> | null = $state(null);
	let Scene: Component<Record<string, unknown>> | null = $state(null);
	let status = $state<'loading' | 'ready' | 'fallback'>('loading');
	let active = $state(true);
	let hovered = $state<HotspotId | null>(null);
	let selected = $state<HotspotId | null>(null);
	let interacted = $state(false);
	const focus = $derived(selected ?? hovered);
	const info = $derived(focus ? HOTSPOTS[focus] : null);

	onMount(() => {
		const io = new IntersectionObserver(([e]) => (active = e.isIntersecting), { rootMargin: '120px' });
		io.observe(host);
		if (!webglAvailable()) {
			status = 'fallback';
			return () => io.disconnect();
		}
		Promise.all([import('@threlte/core'), import('./HeroScene.svelte')])
			.then(([core, scene]) => {
				Canvas = core.Canvas as unknown as Component<Record<string, unknown>>;
				Scene = scene.default as unknown as Component<Record<string, unknown>>;
				status = 'ready';
			})
			.catch(() => (status = 'fallback'));
		return () => io.disconnect();
	});

	function select(id: HotspotId) {
		selected = selected === id ? null : id;
		interacted = true;
	}
</script>

<div class="stage" bind:this={host} class:ready={status === 'ready' && play}>
	<div class="aura" aria-hidden="true"></div>
	<div class="canvas" role="img" aria-label="Interactive 3D model of a ZENRES-connected factory, container, freight loop and export port">
		{#if status === 'ready' && Canvas && Scene}
			<Canvas shadows dpr={[1, 1.75]}>
				<Scene {play} {active} {selected} {hovered} onhover={(id: HotspotId | null) => (hovered = id)} onselect={select} />
			</Canvas>
		{:else if status === 'fallback'}
			<div class="fallback">
				<svg viewBox="0 0 400 300" aria-hidden="true">
					<defs>
						<linearGradient id="hs-g" x1="0" x2="1"><stop offset="0" stop-color="#a4d4f7" /><stop offset="1" stop-color="#cbb5ed" /></linearGradient>
					</defs>
					<path d="M40 200 L200 110 L360 200 L200 290 Z" fill="#f5f8fd" stroke="#dfe7f2" />
					<path d="M90 190 L170 145 L230 178 L150 223 Z" fill="#fbfcff" stroke="#e4ebf5" />
					<rect x="230" y="190" width="60" height="26" rx="5" fill="url(#hs-g)" />
				</svg>
			</div>
		{:else}
			<div class="loading"><span></span></div>
		{/if}
	</div>

	{#if status === 'ready'}
		<div class="hint" class:gone={interacted}>
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><path d="M12 3v4M12 17v4M3 12h4M17 12h4" /><circle cx="12" cy="12" r="3" /></svg>
			Drag to explore · tap a marker
		</div>
	{/if}

	<div class="legend" role="tablist" aria-label="Explore the platform">
		{#each Object.entries(HOTSPOTS) as [id, h] (id)}
			<button
				role="tab"
				aria-selected={focus === id}
				onpointerenter={() => (hovered = id as HotspotId)}
				onpointerleave={() => (hovered = null)}
				onclick={() => select(id as HotspotId)}
			>
				<span class="n">{h.n}</span>{h.pillar.replace('ZENRES ', '')}
			</button>
		{/each}
	</div>

	{#key focus}
		{#if info}
			<aside class="info" aria-live="polite">
				<span class="eyebrow">{info.n} — {info.pillar}</span>
				<h3>{info.title}</h3>
				<p>{info.body}</p>
				<div class="stat"><span class="num">{info.stat[0]}</span><span>{info.stat[1]}</span></div>
				<a href={info.href} class="more">Explore <span aria-hidden="true">→</span></a>
			</aside>
		{/if}
	{/key}
</div>

<style>
	.stage {
		position: relative;
		width: 100%;
		height: 100%;
		min-height: 520px;
		opacity: 0;
		transform: translateY(24px) scale(0.985);
		transition:
			opacity 1.4s var(--ease-expo) 0.35s,
			transform 1.6s var(--ease-expo) 0.35s;
	}
	.stage.ready {
		opacity: 1;
		transform: none;
	}
	.aura {
		position: absolute;
		inset: 8% 4% 10% 6%;
		border-radius: 50%;
		background:
			radial-gradient(closest-side at 38% 46%, rgba(164, 212, 247, 0.55), transparent 70%),
			radial-gradient(closest-side at 66% 58%, rgba(203, 181, 237, 0.5), transparent 72%);
		filter: blur(20px);
		pointer-events: none;
		animation: aura 12s ease-in-out infinite alternate;
	}
	@keyframes aura {
		to {
			transform: translate3d(2%, -2%, 0) scale(1.05);
		}
	}
	.canvas {
		position: absolute;
		inset: 0;
		touch-action: pan-y;
	}
	.canvas :global(canvas) {
		outline: none;
	}
	.loading {
		height: 100%;
		display: grid;
		place-items: center;
	}
	.loading span {
		width: 26px;
		height: 26px;
		border-radius: 50%;
		border: 2px solid var(--border-strong);
		border-top-color: var(--brand-emerald);
		animation: spin 0.9s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	.fallback {
		height: 100%;
		display: grid;
		place-items: center;
	}
	.fallback svg {
		width: 80%;
	}
	.hint {
		position: absolute;
		left: 50%;
		bottom: 64px;
		transform: translateX(-50%);
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 7px 14px;
		border-radius: 999px;
		background: rgba(252, 253, 255, 0.6);
		border: 1px solid rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		font-size: 12px;
		font-weight: 500;
		color: var(--ink-secondary);
		pointer-events: none;
		transition: opacity 600ms var(--ease-out);
		animation: hint 2.4s ease-in-out infinite;
	}
	.hint.gone {
		opacity: 0;
	}
	@keyframes hint {
		50% {
			transform: translateX(-50%) translateY(-3px);
		}
	}
	.legend {
		position: absolute;
		left: 50%;
		bottom: 8px;
		transform: translateX(-50%);
		display: flex;
		gap: 4px;
		padding: 4px;
		border-radius: 999px;
		background: rgba(252, 253, 255, 0.62);
		border: 1px solid rgba(255, 255, 255, 0.9);
		box-shadow: 0 10px 30px -16px rgba(67, 92, 140, 0.45);
		backdrop-filter: blur(14px) saturate(1.4);
		-webkit-backdrop-filter: blur(14px) saturate(1.4);
	}
	.legend button {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		height: 32px;
		padding: 0 12px 0 5px;
		border-radius: 999px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--ink-secondary);
		white-space: nowrap;
		transition: background 300ms, color 300ms;
	}
	.legend .n {
		display: grid;
		place-items: center;
		width: 22px;
		height: 22px;
		border-radius: 50%;
		background: var(--surface-subtle);
		font-size: 10px;
		font-weight: 700;
		transition: background 300ms;
	}
	.legend button[aria-selected='true'] {
		background: #fff;
		color: var(--ink-primary);
		box-shadow: 0 2px 10px -4px rgba(67, 92, 140, 0.4);
	}
	.legend button[aria-selected='true'] .n {
		background: linear-gradient(135deg, #a4d4f7, #cbb5ed);
		color: #32435c;
	}
	.info {
		position: absolute;
		right: 2%;
		top: 4%;
		width: min(300px, 46%);
		padding: 18px 18px 16px;
		border-radius: var(--radius-md);
		background: rgba(252, 253, 255, 0.72);
		border: 1px solid rgba(255, 255, 255, 0.95);
		box-shadow: 0 24px 60px -28px rgba(67, 92, 140, 0.55);
		backdrop-filter: blur(18px) saturate(1.5);
		-webkit-backdrop-filter: blur(18px) saturate(1.5);
		display: flex;
		flex-direction: column;
		gap: 8px;
		animation: info-in 600ms var(--ease-expo);
	}
	@keyframes info-in {
		from {
			opacity: 0;
			transform: translateY(10px) scale(0.98);
		}
	}
	.info h3 {
		font-size: 1.2rem;
		letter-spacing: -0.02em;
	}
	.info p {
		font-size: 13px;
		line-height: 1.55;
	}
	.stat {
		display: flex;
		align-items: baseline;
		gap: 8px;
		padding-top: 8px;
		border-top: 1px solid rgba(189, 203, 220, 0.5);
		font-size: 11.5px;
		color: var(--ink-muted);
	}
	.stat .num {
		font-size: 1.5rem;
		font-weight: 600;
		letter-spacing: -0.03em;
		background: linear-gradient(105deg, var(--brand-emerald), var(--accent-cobalt));
		-webkit-background-clip: text;
		background-clip: text;
		color: transparent;
	}
	.more {
		align-self: flex-start;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--brand-emerald);
	}
	@media (max-width: 760px) {
		.legend button {
			padding: 0 8px 0 4px;
			font-size: 0;
			gap: 0;
		}
		.info {
			left: 4%;
			right: 4%;
			top: auto;
			bottom: 60px;
			width: auto;
		}
		.hint {
			display: none;
		}
	}
</style>
